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

package egress

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type setCommand struct{}

type setOptions struct {
	workloadID  string
	namespaceID string

	bind           string
	hosts          []string
	labelWhitelist []string

	parsedBind string
	parsedPort *v1alpha3.Port

	// TODO 2: configurable bind, port
}

func newSetCommand(cli cli.CLI) *cobra.Command {
	c := &setCommand{}
	options := &setOptions{}

	cmd := &cobra.Command{
		Use:           "set [--namespace] namespace [--workload name] [--bind [PROTOCOL://[IP]:port]|[unix://socket] [--hosts h1,h2]",
		Short:         "Set sidecar egress rule for a workload",
		Args:          cobra.MaximumNArgs(2),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) == 1 {
				if options.namespaceID == "" {
					options.namespaceID = args[0]
				} else if options.workloadID == "" {
					options.workloadID = args[0]
				}
			}
			if len(args) == 2 {
				options.namespaceID = args[0]
				options.workloadID = args[1]
			}

			if options.namespaceID == "" {
				return errors.New("namespace must be specified")
			}

			if options.workloadID != "" {
				_, err = util.ParseK8sResourceID(options.namespaceID + "/" + options.workloadID)
				if err != nil {
					return errors.WrapIf(err, "could not parse workload ID")
				}
			}

			options.parsedBind, options.parsedPort, err = common.ParseSidecarEgressBind(options.bind)
			if err != nil {
				return errors.WrapIf(err, "could not parse bind option")
			}

			err = c.validateOptions(options)
			if err != nil {
				return errors.WithStack(err)
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.namespaceID, "namespace", "", "Workload name")
	flags.StringVar(&options.workloadID, "workload", "", "Namespace name")
	flags.StringVarP(&options.bind, "bind", "b", "", "Egress listener bind [PROTOCOL://[IP]:port]|[unix://socket]")
	flags.StringArrayVar(&options.hosts, "hosts", options.hosts, "Egress listener Hosts")
	flags.StringArrayVarP(&options.labelWhitelist, "labelWhitelist", "l", options.labelWhitelist, "Labels to include in the workload selector")

	return cmd
}

func (c *setCommand) validateOptions(options *setOptions) error {
	// TODO 2: configurable port, bind -> len(hosts)!=0 is conditional
	if len(options.hosts) == 0 {
		return errors.New("at least one host must be specified")
	}

	return nil
}

func (c *setCommand) run(cli cli.CLI, options *setOptions) error {
	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	response, err := applyEgress(client, options.namespaceID, options.workloadID, options.parsedBind, options.hosts, options.parsedPort, options.labelWhitelist)
	if err != nil {
		return errors.WrapIf(err, "could not apply sidecar egress rules")
	}

	if !response {
		return errors.New("unknown internal error: could not apply sidecar egress")
	}

	var sidecars []graphql.Sidecar
	if options.workloadID != "" {
		workload, err := client.GetWorkloadWithSidecar(options.namespaceID, options.workloadID)
		if err != nil {
			return errors.Wrap(err, "couldn't query workload sidecars")
		}
		sidecars = workload.Sidecars
	} else {
		resp, err := client.GetNamespaceWithSidecar(options.namespaceID)
		if err != nil {
			return errors.Wrap(err, "couldn't query namespace sidecars")
		}
		sidecars = resp.Namespace.Sidecars
	}

	log.Infof("sidecar egress for %s/%s set successfully\n\n", options.namespaceID, options.workloadID)

	return Output(cli, options.namespaceID, options.workloadID, sidecars, false, false)
}

func applyEgress(client graphql.Client, namespace, name, bind string, hosts []string, port *v1alpha3.Port, labelWhitelist []string) (bool, error) {
	req := graphql.ApplySidecarEgressInput{
		Selector: graphql.SidecarEgressSelector{
			Namespace: namespace,
		},
		Egress: graphql.Egress{
			Hosts: hosts,
		},
	}

	if name != "" {
		workload, err := client.GetWorkloadWithSidecar(namespace, name)
		if err != nil {
			return false, errors.WrapIf(err, "could not find workload in mesh, check the workload ID")
		}
		if len(labelWhitelist) > 0 {
			labels := make(map[string]string)
			for _, l := range labelWhitelist {
				if wl, ok := workload.Labels[l]; ok {
					labels[l] = wl
				}
			}
			if len(labels) == 0 {
				return false, errors.New("workload has no matching label from label whitelist")
			}
			req.Selector.WorkloadLabels = &labels
		} else {
			req.Selector.WorkloadLabels = &workload.Labels
		}
	}

	if bind != "" {
		req.Selector.Bind = &bind
	}

	if port != nil {
		req.Selector.Port = port
	}

	response, err := client.ApplySidecarEgress(req)
	if err != nil {
		return false, errors.WrapIf(err, "could not apply sidecar egress")
	}
	return bool(response), nil
}
