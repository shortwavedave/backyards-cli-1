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
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/square/go-jose/v3/jwt"
	"k8s.io/client-go/util/homedir"

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

	GetPersistentConfig() PersistentConfig
	GetPersistentGlobalConfig() PersistentGlobalConfig
	GetToken() string

	LabelManager() k8s.LabelManager

	// An endpoint can be currently:
	// - external HTTP(s) endpoint
	// - local port-forward to an HTTP(s) endpoint (deprecated)
	// - local HTTP(s) proxy
	InitializedEndpoint() (endpoint.Endpoint, error)
	PersistentEndpoint() (endpoint.Endpoint, error)

	Initialize() error

	IfConfirmed(string, func() error) error
}

type backyardsCLI struct {
	out                    io.Writer
	rootCmd                *cobra.Command
	labelManager           k8s.LabelManager
	lmOnce                 sync.Once
	persistentConfig       PersistentConfig
	persistentGlobalConfig PersistentGlobalConfig
}

func NewCli(out io.Writer, rootCmd *cobra.Command) CLI {
	return &backyardsCLI{
		out:     out,
		lmOnce:  sync.Once{},
		rootCmd: rootCmd,
	}
}

func (c *backyardsCLI) Initialize() error {
	err := c.loadPersistentConfig()
	if err != nil {
		return err
	}
	persistentConfigExists, err := fileExists(c.persistentConfig.GetConfigFileUsed())
	if err != nil {
		return err
	}
	if !persistentConfigExists {
		if viper.GetString("kubeconfig") == "" && viper.GetString("kubecontext") == "" {
			config, err := getValidatedRawConfig()
			if err != nil {
				return errors.WrapIf(err, "failed to get raw kubernetes config")
			}
			message := fmt.Sprintf("Are you sure to use the current context? %s (API Server: %s)",
				config.CurrentContext, config.Clusters[config.Contexts[config.CurrentContext].Cluster].Server)
			confirmed, err := c.confirm(message, true)
			if err != nil {
				return err
			}
			if !confirmed {
				return errors.New("refusing to use current context")
			}
		}
	}
	err = c.loadPersistentGlobalConfig()
	if err != nil {
		return err
	}
	return nil
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

func (c *backyardsCLI) getPortforwardForIGW(localPort int) (*portforward.Portforward, error) {
	return c.getPortforwardForPod(IGWMatchLabels, c.persistentConfig.Namespace(), localPort, IGWPort)
}

func (c *backyardsCLI) getPortforwardForPod(podLabels map[string]string, namespace string, localPort, remotePort int) (*portforward.Portforward, error) {
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

func (c *backyardsCLI) GetPersistentConfig() PersistentConfig {
	return c.persistentConfig
}

func (c *backyardsCLI) loadPersistentGlobalConfig() error {
	v, err := createViper(filepath.Join(homedir.HomeDir(), ".banzai/backyards/config.yaml"))
	if err != nil {
		return err
	}

	PersistentGlobalSettings.Bind(v, c.rootCmd.Flags())
	c.persistentGlobalConfig = newViperPersistentGlobalConfig(v)
	return err
}

func (c *backyardsCLI) loadPersistentConfig() error {
	configFile, err := c.persistentConfigFile()
	if err != nil {
		return err
	}

	v, err := createViper(configFile)
	if err != nil {
		return err
	}
	PersistentSettings.Bind(v, c.rootCmd.Flags())
	c.persistentConfig = newViperPersistentConfig(v)
	return nil
}

func (c *backyardsCLI) persistentConfigFile() (string, error) {
	if viper.GetString(PersistentConfigKey) != "" {
		return viper.GetString(PersistentConfigKey), nil
	}

	config, err := getValidatedRawConfig()
	if err != nil {
		return "", err
	}

	parse, err := url.Parse(config.Clusters[config.Contexts[config.CurrentContext].Cluster].Server)
	if err != nil {
		return "", errors.WrapIf(err, "failed to parse server url")
	}

	host := parse.Host
	host = strings.Replace(host, ":", "_", -1)
	host = strings.Replace(host, ".", "_", -1)

	configFileName := fmt.Sprintf("%s@%s", config.CurrentContext, host)

	return filepath.Join(homedir.HomeDir(), ".banzai/backyards", configFileName+".yaml"), nil
}

func (c *backyardsCLI) GetK8sClient() (k8sclient.Client, error) {
	config, err := c.GetK8sConfig()
	if err != nil {
		return nil, err
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

func (c *backyardsCLI) withHealthCheck(ep endpoint.Endpoint) (endpoint.Endpoint, error) {
	client := retryablehttp.NewClient()
	client.RetryWaitMin = time.Millisecond * 50
	client.RetryWaitMax = time.Millisecond * 100
	client.RetryMax = 5
	client.Logger = hclog.NewNullLogger()
	level := hclog.Info
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		level = hclog.Debug
	}
	client.Logger = hclog.New(&hclog.LoggerOptions{
		Name:   "health check",
		Level:  level,
		Output: c.rootCmd.ErrOrStderr(),
	})
	client.HTTPClient = ep.HTTPClient()
	_, err := client.Get(ep.URLForPath(""))
	if err != nil {
		return nil, errors.WrapIf(err, "failed to health check the created endpoint")
	}
	return ep, nil
}

func (c *backyardsCLI) endpoint(persistentPort int) (endpoint.Endpoint, error) {
	url := c.persistentConfig.BaseURL()
	ca, err := c.getEndpointCA()
	if err != nil {
		return nil, err
	}
	if url == "" {
		cfg, err := c.GetK8sConfig()
		if err != nil {
			return nil, err
		}

		port := c.persistentConfig.LocalPort()
		if port == -1 {
			port = persistentPort
		}

		if c.persistentConfig.UsePortForward() {
			pf, err := c.getPortforwardForIGW(port)
			if err != nil {
				return nil, err
			}

			err = pf.Run()
			if err != nil {
				return nil, err
			}
			return c.withHealthCheck(endpoint.NewPortforwardEndpoint(pf, ca))
		}

		ep, err := endpoint.NewProxyEndpoint(port, cfg, endpoint.K8sService{
			Name:      BackyardsIngressServiceName,
			Namespace: c.persistentConfig.Namespace(),
			Port:      80,
		})
		if err != nil {
			return nil, err
		}
		return c.withHealthCheck(ep)
	}

	return c.withHealthCheck(endpoint.NewExternalEndpoint(url, ca))
}

func (c *backyardsCLI) getEndpointCA() ([]byte, error) {
	if c.persistentConfig.CACert() != "" {
		return ioutil.ReadFile(c.persistentConfig.CACert())
	}
	return nil, nil
}

func (c *backyardsCLI) GetToken() string {
	token := c.persistentConfig.Token()
	if token != "" {
		claims := &jwt.Claims{}
		webToken, err := jwt.ParseSigned(token)
		if err != nil {
			logrus.Error("Failed to parse signed token", err)
			return ""
		}
		if err := webToken.UnsafeClaimsWithoutVerification(claims); err != nil {
			logrus.Error("Failed to extract claims from token", err)
			return ""
		}
		if err := claims.ValidateWithLeeway(jwt.Expected{}.WithTime(time.Now()), 0); err != nil {
			logrus.Debug("Token expired")
			return ""
		}
	}
	return token
}

func (c *backyardsCLI) GetPersistentGlobalConfig() PersistentGlobalConfig {
	return c.persistentGlobalConfig
}

func (c *backyardsCLI) IfConfirmed(message string, action func() error) error {
	confirmed, err := c.confirm(message, false)
	if err != nil {
		return err
	}
	if confirmed {
		return action()
	}
	return nil
}

func (c *backyardsCLI) confirm(message string, defaultSelect bool) (bool, error) {
	confirmed := false
	if c.InteractiveTerminal() {
		err := survey.AskOne(&survey.Confirm{
			Renderer: survey.Renderer{},
			Message:  message,
			Default:  defaultSelect,
		}, &confirmed)
		if err != nil {
			return false, errors.Wrapf(err, "failed to get confirmation")
		}
	}
	return confirmed, nil
}
