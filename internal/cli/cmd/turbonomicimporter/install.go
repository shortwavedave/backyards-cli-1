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

package turbonomicimporter

import (
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/turbonomicimporter"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

type installCommand struct {
	cli cli.CLI
}

type InstallOptions struct {
	namespace string

	DumpResources bool

	TurbonomicHostname           string
	TurbonomicUsername           string
	TurbonomicPassword           string
	TurbonomicInsecureSkipVerify bool
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{}
}

func NewInstallCommand(cli cli.CLI, options *InstallOptions) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:     "install [flags]",
		Args:    cobra.NoArgs,
		Short:   "Install turbonomic importer",
		Long:    `Installs turbonomic importer`,
		Example: `backyards turbonomic-importer install`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if options.namespace == "" {
				options.namespace = turbonomicImporterNamespace
			}

			return c.run(cli, options)
		},
	}

	cmd.Flags().BoolVarP(&options.DumpResources, "dump-resources", "d", options.DumpResources, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *InstallOptions) error {
	objects, err := getTurbonomicImporterObjects(options.namespace, options.TurbonomicHostname,
		options.TurbonomicUsername, options.TurbonomicPassword, options.TurbonomicInsecureSkipVerify)
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

func getTurbonomicImporterObjects(namespace, hostname, username, password string, insecureSkipVerify bool) (object.K8sObjects, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(turbonomicimporter.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	values.SetDefaults(hostname, username, password, insecureSkipVerify)

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(turbonomicimporter.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "turbonomic-importer",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	}, "turbonomic-importer")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}
