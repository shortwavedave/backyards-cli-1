// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"

	"github.com/banzaicloud/backyards-cli/internal/endpoint"
	internalk8s "github.com/banzaicloud/backyards-cli/internal/k8s"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
	"github.com/banzaicloud/backyards-cli/pkg/k8s/portforward"
	"github.com/banzaicloud/backyards-cli/pkg/output"
)

var (
	defaultLocalEndpointPort = 50500
	IGWPort                  = 80
	IGWMatchLabels           = map[string]string{
		"app.kubernetes.io/component": "ingressgateway",
		"app.kubernetes.io/instance":  "backyards",
	}
	BackyardsServiceAccountName = "backyards"
	BackyardsIngressServiceName = "backyards-ingressgateway"
)

type CLI interface {
	Out() io.Writer
	OutputFormat() string
	Color() bool
	Interactive() bool
	InteractiveTerminal() bool
	GetRootCommand() *cobra.Command
	GetK8sClient() (k8sclient.Client, error)
	GetK8sConfig() (*rest.Config, error)
	LabelManager() k8s.LabelManager

	// An endpoint can be currently:
	// - external HTTP(s) endpoint
	// - local port-forward to an HTTP(s) endpoint
	// - planned: local HTTP(s) proxy
	InitializedEndpoint() (endpoint.Endpoint, error)
	PersistentEndpoint() (endpoint.Endpoint, error)

	Stop() error
}

type backyardsCLI struct {
	out          io.Writer
	rootCmd      *cobra.Command
	labelManager k8s.LabelManager
	lmOnce       sync.Once
}

func NewCli(out io.Writer, rootCmd *cobra.Command) CLI {
	return &backyardsCLI{
		out:     out,
		lmOnce:  sync.Once{},
		rootCmd: rootCmd,
	}
}

func (c *backyardsCLI) GetRootCommand() *cobra.Command {
	return c.rootCmd
}

func (c *backyardsCLI) InteractiveTerminal() bool {
	return c.Interactive() && c.OutputFormat() == output.OutputFormatTable
}

func (c *backyardsCLI) Interactive() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd()) {
		return !viper.GetBool("formatting.non-interactive")
	}

	return viper.GetBool("formatting.force-interactive")
}

func (c *backyardsCLI) Color() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return !viper.GetBool("formatting.no-color")
	}

	return viper.GetBool("formatting.force-color")
}

func (c *backyardsCLI) OutputFormat() string {
	return viper.GetString("output.format")
}

func (c *backyardsCLI) Out() io.Writer {
	return c.out
}

func (c *backyardsCLI) GetPortforwardForIGW(localPort int) (*portforward.Portforward, error) {
	return c.GetPortforwardForPod(IGWMatchLabels, viper.GetString("backyards.namespace"), localPort, IGWPort)
}

func (c *backyardsCLI) GetPortforwardForPod(podLabels map[string]string, namespace string, localPort, remotePort int) (*portforward.Portforward, error) {
	config, err := c.GetK8sConfig()
	if err != nil {
		return nil, err
	}

	client, err := c.GetK8sClient()
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Creating port forward: local port %d namespace: %s pod labels: %s remote port: %d",
		localPort, namespace, podLabels, remotePort)
	pf, err := portforward.New(client, config, podLabels, namespace, localPort, remotePort)
	if err != nil {
		return nil, err
	}

	return pf, nil
}

func (c *backyardsCLI) GetK8sClient() (k8sclient.Client, error) {
	config, err := k8sclient.GetConfigWithContext(viper.GetString("kubeconfig"), viper.GetString("kubecontext"))
	if err != nil {
		return nil, errors.WrapIf(err, "could not get k8s config")
	}

	client, err := k8sclient.NewClient(config, k8sclient.Options{})
	if err != nil {
		return nil, errors.WrapIf(err, "could not get k8s client")
	}

	return client, nil
}

func (c *backyardsCLI) GetK8sConfig() (*rest.Config, error) {
	config, err := k8sclient.GetConfigWithContext(viper.GetString("kubeconfig"), viper.GetString("kubecontext"))
	if err != nil {
		return nil, errors.WrapIf(err, "could not get k8s config")
	}

	return config, nil
}

func (c *backyardsCLI) LabelManager() k8s.LabelManager {
	c.lmOnce.Do(func() {
		c.labelManager = internalk8s.NewLabelManager(c.InteractiveTerminal(), c.GetRootCommand().Version)
	})
	return c.labelManager
}

func (c *backyardsCLI) InitializedEndpoint() (endpoint.Endpoint, error) {
	return c.endpoint(0)
}

func (c *backyardsCLI) PersistentEndpoint() (endpoint.Endpoint, error) {
	return c.endpoint(defaultLocalEndpointPort)
}

func withHealthCheck(ep endpoint.Endpoint) (endpoint.Endpoint, error) {
	client := retryablehttp.NewClient()
	client.RetryWaitMin = time.Millisecond * 50
	client.RetryWaitMax = time.Millisecond * 100
	client.RetryMax = 5
	client.Logger = hclog.NewNullLogger()
	client.HTTPClient = ep.HTTPClient()
	_, err := client.Get(ep.URLForPath("/"))
	if err != nil {
		return nil, errors.WrapIf(err, "failed to health check the created endpoint")
	}
	return ep, nil
}

func (c *backyardsCLI) endpoint(persistentPort int) (endpoint.Endpoint, error) {
	url := viper.GetString("backyards.url")
	ca, err := getEndpointCA()
	if err != nil {
		return nil, err
	}
	if url == "" {
		cfg, err := c.GetK8sConfig()
		if err != nil {
			return nil, err
		}

		port := viper.GetInt("backyards.localPort")
		if port == -1 {
			port = persistentPort
		}

		if viper.GetBool("backyards.usePortforward") {
			pf, err := c.GetPortforwardForIGW(port)
			if err != nil {
				return nil, err
			}

			err = pf.Run()
			if err != nil {
				return nil, err
			}
			return withHealthCheck(endpoint.NewPortforwardEndpoint(pf, ca))
		}

		ep, err := endpoint.NewProxyEndpoint(port, cfg, endpoint.K8sService{
			Name:      BackyardsIngressServiceName,
			Namespace: viper.GetString("backyards.namespace"),
			Port:      80,
		})
		if err != nil {
			return nil, err
		}
		return withHealthCheck(ep)
	}

	return withHealthCheck(endpoint.NewExternalEndpoint(url, ca))
}

func getEndpointCA() ([]byte, error) {
	if viper.GetString("backyards.cacert") != "" {
		return ioutil.ReadFile(viper.GetString("backyards.cacert"))
	}
	return nil, nil
}

func (c *backyardsCLI) Stop() error {
	return nil
}
