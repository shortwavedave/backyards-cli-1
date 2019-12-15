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

package kafka

import (
	"fmt"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/zookeeper"
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
		Use:   "uninstall [flags]",
		Args:  cobra.NoArgs,
		Short: "Output or delete Kubernetes resources to uninstall kafka cluster",
		Long: `Output or delete Kubernetes resources to uninstall kafka cluster.

The command automatically removes the resources.
It can only dump the removable resources with the '--dump-resources' option.`,
		Example: `  # Default uninstall.
  backyards canary uninstall

  # Uninstall canary feature from a non-default namespace.
  backyards canary uninstall install -n custom-istio-ns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if options.namespace == "" {
				options.namespace = kafkaNamespace
			}

			if options.UninstallEverything {
				return c.run(cli, options)
			}

			return cli.IfConfirmed("Uninstall kafka cluster. This command will destroy resources and cannot be undone. Are you sure to proceed?", func() error {
				return c.run(cli, options)
			})
		},
	}

	cmd.Flags().BoolVarP(&options.DumpResources, "dump-resources", "d", options.DumpResources, "Dump resources to stdout instead of applying them")
	cmd.Flags().BoolVarP(&options.UninstallEverything, "uninstall-everything", "a", options.UninstallEverything, "Uninstall all components at once")

	return cmd
}

func (c *uninstallCommand) run(cli cli.CLI, options *UninstallOptions) error {
	err := c.uninstallKafka(cli, options)
	if err != nil {
		return err
	}

	err = c.uninstallZookeeper(cli, options)
	if err != nil {
		return err
	}

	err = c.uninstallCertManager(cli, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *uninstallCommand) uninstallCertManager(cli cli.CLI, options *UninstallOptions) error {
	var err error
	var scmd *cobra.Command

	scmdOptions := certmanager.NewUninstallOptions()
	if options.DumpResources {
		scmdOptions.DumpResources = true
	}
	if options.UninstallEverything {
		scmdOptions.UninstallEverything = true
	}
	scmd = certmanager.NewUninstallCommand(cli, scmdOptions)
	err = scmd.RunE(scmd, nil)
	if err != nil {
		return errors.WrapIf(err, "error during cert-manager uninstall")
	}

	return nil
}

func (c *uninstallCommand) uninstallZookeeper(cli cli.CLI, options *UninstallOptions) error {
	var err error
	var scmd *cobra.Command

	scmdOptions := zookeeper.NewUninstallOptions()
	if options.DumpResources {
		scmdOptions.DumpResources = true
	}
	if options.UninstallEverything {
		scmdOptions.UninstallEverything = true
	}
	scmd = zookeeper.NewUninstallCommand(cli, scmdOptions)
	err = scmd.RunE(scmd, nil)
	if err != nil {
		return errors.WrapIf(err, "error during zookeeper uninstall")
	}

	return nil
}

func (c *uninstallCommand) uninstallKafka(cli cli.CLI, options *UninstallOptions) error {
	objects, err := getK8sObjects(options.namespace)
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

	err = k8s.DeleteResources(client, c.cli.LabelManager(), objects, k8s.WaitForResourceConditions(backoff, k8s.NonExistsConditionCheck))
	if err != nil {
		return errors.WrapIf(err, "could not delete k8s resources")
	}

	return nil
}
