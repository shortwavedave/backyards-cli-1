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

	"emperror.dev/errors"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/apis/istio/v1alpha3"

	internalk8s "github.com/banzaicloud/backyards-cli/internal/k8s"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"

	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
	"github.com/banzaicloud/backyards-cli/pkg/k8s/portforward"
	"github.com/banzaicloud/backyards-cli/pkg/output"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

var (
	defaultPortForward = 50500
	IGWPort            = 80
	IGWMatchLabels     = map[string]string{
		"app.kubernetes.io/component": "ingressgateway",
		"app.kubernetes.io/instance":  "backyards",
	}
	BackyardsServiceAccountName = "backyards"
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
	InitializedEndpoint() (Endpoint, error)
	PersistentEndpoint() (Endpoint, error)

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

	err = istiov1beta1.AddToScheme(k8sclient.GetScheme())
	if err != nil {
		return nil, errors.WrapIf(err, "could not add istio-operator/v1beta1 to scheme")
	}
	err = apiextensionsv1beta1.AddToScheme(k8sclient.GetScheme())
	if err != nil {
		return nil, errors.WrapIf(err, "could not add apiextensions/v1beta1 to scheme")
	}
	err = v1alpha3.AddToScheme(k8sclient.GetScheme())
	if err != nil {
		return nil, errors.WrapIf(err, "could not add istio/v1alpha3 to scheme")
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

func (c *backyardsCLI) InitializedEndpoint() (Endpoint, error) {
	return c.endpoint(0)
}

func (c *backyardsCLI) PersistentEndpoint() (Endpoint, error) {
	return c.endpoint(defaultPortForward)
}

func (c *backyardsCLI) endpoint(persistentPort int) (Endpoint, error) {
	url := viper.GetString("backyards.url")
	ca, err := getEndpointCA()
	if err != nil {
		return nil, err
	}
	if url == "" {
		port := viper.GetInt("backyards.portforward")
		if port == 0 && persistentPort > 0 {
			port = persistentPort
		}
		pf, err := c.GetPortforwardForIGW(port)
		if err != nil {
			return nil, err
		}
		err = pf.Run()
		if err != nil {
			return nil, err
		}
		return &portForwardEndpoint{
			pf: pf,
			ca: ca,
		}, nil
	}
	return &externalEndpoint{
		baseURL: url,
		ca:      ca,
	}, nil
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
