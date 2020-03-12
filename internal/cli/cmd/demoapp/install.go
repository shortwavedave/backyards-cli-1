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

package demoapp

import (
	"context"
	"fmt"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/backyards_demo"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

const (
	istioNotFoundErrorTemplate = `Unable to install Backyards: %s

An existing Istio installation is required. You can install it with:

backyards istio install
`
)

var availableServices = map[string]bool{
	"bombardier":    true,
	"analytics":     true,
	"bookings":      true,
	"catalog":       true,
	"frontpage":     true,
	"movies":        true,
	"notifications": true,
	"payments":      true,
	"database":      true,
}

type installCommand struct {
	cli cli.CLI
}

type InstallOptions struct {
	namespace       string
	istioNamespace  string
	peerCluster     bool
	enabledServices []string

	DumpResources bool
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{}
}

func NewInstallCommand(cli cli.CLI, options *InstallOptions) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Install demo application",
		Long: `Installs demo application.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.`,
		Example: `  # Default install.
  backyards demoapp install

  # Install Backyards into a non-default namespace.
  backyards demoapp install -n backyards-system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if options.namespace == "" {
				options.namespace = backyardsDemoNamespace
			}

			if len(options.enabledServices) > 0 {
				for _, n := range options.enabledServices {
					if !availableServices[n] {
						return errors.Errorf("invalid service '%s'", n)
					}
				}
			}

			return c.run(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", "istio-system", "Namespace of Istio sidecar injector")
	cmd.Flags().BoolVar(&options.peerCluster, "peer", options.peerCluster, "The destination cluster is a peer in a multi-cluster mesh")
	cmd.Flags().StringSliceVarP(&options.enabledServices, "enabled-services", "s", options.enabledServices, "Enabled services of the demo app")
	cmd.Flags().BoolVarP(&options.DumpResources, "dump-resources", "d", options.DumpResources, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *InstallOptions) error {
	err := c.validate(options.istioNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, istioNotFoundErrorTemplate, err)
		return nil
	}

	objects, err := getBackyardsDemoObjects(options.namespace, options.peerCluster, options.enabledServices...)
	if err != nil {
		return err
	}
	objects.Sort(helm.InstallObjectOrder())

	if !options.DumpResources {
		client, err := cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.ApplyResources(client, cli.LabelManager(), objects)
		if err != nil {
			return err
		}

		err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(objects), wait.Backoff{
			Duration: time.Second * 5,
			Factor:   1,
			Jitter:   0,
			Steps:    24,
		}, k8s.ExistsConditionCheck, k8s.ReadyReplicasConditionCheck)
		if err != nil {
			return err
		}
	} else {
		yaml, err := objects.YAMLManifest()
		if err != nil {
			return err
		}
		_, err = c.cli.Out().Write([]byte(yaml))
		if err != nil {
			return err
		}
	}

	return nil
}

func getBackyardsDemoObjects(namespace string, peerCluster bool, enabledServices ...string) (object.K8sObjects, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(backyards_demo.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	values.UseNamespaceResource = true
	if peerCluster {
		values.IstioResources = false
	}
	setEnabledServices(&values, enabledServices)

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(backyards_demo.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "backyards-demo",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	}, "backyards-demo")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}

func setEnabledServices(values *Values, enabledServices []string) {
	if len(enabledServices) == 0 {
		return
	}
	services := make(map[string]bool)
	for _, n := range enabledServices {
		services[n] = true
	}

	values.Analytics = false
	if services["analytics"] {
		values.Analytics = true
	}
	values.Bookings = false
	if services["bookings"] {
		values.Bookings = true
	}
	values.Catalog = false
	if services["catalog"] {
		values.Catalog = true
	}
	values.Frontpage = false
	if services["frontpage"] {
		values.Frontpage = true
	}
	values.MoviesV1 = false
	values.MoviesV2 = false
	values.MoviesV3 = false
	if services["movies"] {
		values.MoviesV1 = true
		values.MoviesV2 = true
		values.MoviesV3 = true
	}
	values.Notifications = false
	if services["notifications"] {
		values.Notifications = true
	}
	values.Payments = false
	if services["payments"] {
		values.Payments = true
	}
	values.Database = false
	if services["database"] {
		values.Database = true
	}
	values.Bombardier = false
	if services["bombardier"] {
		values.Bombardier = true
	}
}

func (c *installCommand) validate(istioNamespace string) error {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		return errors.WrapIf(err, "could not get k8s client")
	}
	var pods v1.PodList
	err = cl.List(context.Background(), &pods, client.InNamespace(istioNamespace), client.MatchingLabels(util.SidecarPodLabels))
	if err != nil {
		return errors.WrapIf(err, "could not list pods")
	}
	if len(pods.Items) == 0 {
		err = cl.List(context.Background(), &pods, client.InNamespace(istioNamespace), client.MatchingLabels(util.IstiodSidecarPodLabels))
		if err != nil {
			return errors.WrapIf(err, "could not list pods")
		}
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			return nil
		}
	}

	if len(pods.Items) > 0 {
		return errors.Errorf("Istio sidecar injector not healthy yet in '%s'", istioNamespace)
	}

	return errors.Errorf("could not find Istio sidecar injector in '%s'", istioNamespace)
}
