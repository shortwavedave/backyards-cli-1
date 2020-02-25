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

package cmd

import (
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/canary"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/turbonomicimporter"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
	"github.com/banzaicloud/backyards-cli/pkg/k8s/resourcemanager"
	"github.com/banzaicloud/backyards-cli/pkg/nodeexporter"
)

type uninstallCommand struct {
	cli cli.CLI
}

type UninstallOptions struct {
	releaseName    string
	istioNamespace string
	dumpResources  bool

	uninstallEverything bool
}

func NewUninstallCommand(cli cli.CLI) *cobra.Command {
	c := &uninstallCommand{
		cli: cli,
	}
	options := &UninstallOptions{}

	cmd := &cobra.Command{
		Use:         "uninstall [flags]",
		Args:        cobra.NoArgs,
		Short:       "Uninstall Backyards",
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.InstallCommand},
		Long: `Uninstall Backyards

The command automatically removes the resources.
It can only dump the removable resources with the '--dump-resources' option.`,
		Example: `  # Default uninstall
  backyards uninstall

  # Uninstall Backyards from a non-default namespace
  backyards uninstall -n backyards-system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			const (
				AnswerAll           = "Remove all resources, including Istio"
				AnswerBackyardsOnly = "Remove Backyards only"
			)

			return cli.IfConfirmed("Uninstall Backyards. This command will destroy resources and cannot be undone. Are you sure to proceed?", func() error {
				if cli.InteractiveTerminal() && !options.uninstallEverything {
					var response string
					fmt.Fprintln(cli.Out(), heredoc.Doc(`
						Do you want to remove all resources deployed by the CLI, or just the Backyards component?
					`))
					err := survey.AskOne(&survey.Select{
						Renderer: survey.Renderer{},
						Default:  AnswerAll,
						Options:  []string{AnswerAll, AnswerBackyardsOnly},
					}, &response)
					if err != nil {
						return err
					}
					options.uninstallEverything = response == AnswerAll
				}
				err := c.run(options)
				if err != nil {
					return err
				}

				err = c.runNodeExporterUninstall(options)
				if err != nil {
					return err
				}

				cli.GetPersistentConfig().SetToken("")
				return c.runSubcommands(options)
			})
		},
	}

	cmd.Flags().StringVar(&options.releaseName, "release-name", "backyards", "Name of the release")
	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", "istio-system", "Namespace of Istio sidecar injector")
	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", false, "Dump resources to stdout instead of applying them")

	cmd.Flags().BoolVarP(&options.uninstallEverything, "uninstall-everything", "a", false, "Uninstall all components at once")

	return cmd
}

func (c *uninstallCommand) run(options *UninstallOptions) error {
	values, err := getValues(options.releaseName, options.istioNamespace, nil)
	if err != nil {
		return err
	}

	objects, err := getBackyardsObjects(values, c.cli)
	if err != nil {
		return err
	}

	objects.Sort(helm.UninstallObjectOrder())

	if !options.dumpResources {
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

	yaml, err := objects.YAMLManifest()
	if err != nil {
		return errors.WrapIf(err, "could not render YAML manifest")
	}
	fmt.Fprint(c.cli.Out(), yaml)

	return nil
}

func (c *uninstallCommand) runSubcommands(options *UninstallOptions) error {
	var err error
	var scmd *cobra.Command

	if options.uninstallEverything {
		scmdOptions := demoapp.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		if options.uninstallEverything {
			scmdOptions.UninstallEverything = true
		}
		scmd = demoapp.NewUninstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during demo application uninstall")
		}
	}

	if options.uninstallEverything {
		scmdOptions := canary.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		if options.uninstallEverything {
			scmdOptions.UninstallEverything = true
		}
		scmd = canary.NewUninstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Canary feature uninstall")
		}
	}

	if options.uninstallEverything {
		scmdOptions := certmanager.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		if options.uninstallEverything {
			scmdOptions.UninstallEverything = true
		}
		scmd = certmanager.NewUninstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during cert-manager uninstall")
		}
	}

	if options.uninstallEverything {
		scmdOptions := istio.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		if options.uninstallEverything {
			scmdOptions.UninstallEverything = true
		}
		scmd = istio.NewUninstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Istio mesh uninstall")
		}
	}

	if options.uninstallEverything {
		scmdOptions := turbonomicimporter.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = turbonomicimporter.NewUninstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during turbonomic-importer uninstall")
		}
	}

	return nil
}

func (c *uninstallCommand) runNodeExporterUninstall(options *UninstallOptions) error {
	client, err := c.cli.GetK8sClient()
	if err != nil {
		return err
	}

	m, err := nodeexporter.NewNodeExporterManager(resourcemanager.New(client, c.cli.LabelManager()), c.cli.GetPersistentConfig().Namespace())
	if err != nil {
		return err
	}
	if options.dumpResources {
		yaml, err := m.Uninstall().YAML()
		if err != nil {
			return err
		}
		_, err = c.cli.Out().Write([]byte(yaml))
		if err != nil {
			return err
		}
	} else {
		err = m.Uninstall().Do()
		if err != nil {
			return err
		}
	}

	return nil
}
