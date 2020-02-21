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
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

type uninstallCommand struct {
	cli cli.CLI
}

type UninstallOptions struct {
	namespace string

	DumpResources       bool
	UninstallEverything bool
}

func NewUninstallOptions() *UninstallOptions {
	return &UninstallOptions{}
}

func NewUninstallCommand(cli cli.CLI, options *UninstallOptions) *cobra.Command {
	c := &uninstallCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:     "uninstall [flags]",
		Args:    cobra.NoArgs,
		Short:   "Output or delete Kubernetes resources to uninstall turbonomic importer",
		Long:    `Output or delete Kubernetes resources to uninstall turbonomic importer`,
		Example: `backyards turbonomic-importer uninstall`,
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

func (c *uninstallCommand) run(cli cli.CLI, options *UninstallOptions) error {
	objects, err := getTurbonomicImporterObjects(options.namespace, "", "", "", false)
	if err != nil {
		return err
	}
	objects.Sort(helm.UninstallObjectOrder())

	if !options.DumpResources {
		err := c.deleteResources(objects)
		if err != nil {
			return errors.WrapIf(err, "could not delete k8s resources")
		}
		return nil
	}

	yaml, err := objects.YAMLManifest()
	if err != nil {
		return errors.WrapIf(err, "could not render YAML manifest")
	}
	fmt.Fprint(cli.Out(), yaml)

	return nil
}

func (c *uninstallCommand) deleteResources(objects object.K8sObjects) error {
	client, err := c.cli.GetK8sClient()
	if err != nil {
		return err
	}

	err = k8s.DeleteResources(client, c.cli.LabelManager(), objects, k8s.WaitForResourceConditions(wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    24,
	}, k8s.NonExistsConditionCheck))
	if err != nil {
		return errors.WrapIf(err, "could not delete k8s resources")
	}

	return nil
}
