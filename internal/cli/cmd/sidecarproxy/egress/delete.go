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

type deleteCommand struct{}

type deleteOptions struct {
	workloadID  string
	namespaceID string

	bind string

	parsedBind string
	parsedPort *v1alpha3.Port
}

func newDeleteCommand(cli cli.CLI) *cobra.Command {
	d := &deleteCommand{}
	options := &deleteOptions{}

	cmd := &cobra.Command{
		Use:           "delete [--namespace] namespace [--workload name] [--bind [PROTOCOL://[IP]:port]|[unix://socket]",
		Short:         "Delete sidecar egress rule for a workload, or for a namespace",
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
				_, err := util.ParseK8sResourceID(options.namespaceID + "/" + options.workloadID)
				if err != nil {
					return errors.WrapIf(err, "could not parse workload ID")
				}
			}

			options.parsedBind, options.parsedPort, err = common.ParseSidecarEgressBind(options.bind)
			if err != nil {
				return errors.WrapIf(err, "could not parse bind option")
			}

			err = d.validateOptions(options)
			if err != nil {
				return errors.WithStack(err)
			}

			return d.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.namespaceID, "namespace", "", "Namespace name")
	flags.StringVar(&options.workloadID, "workload", "", "Workload name")
	flags.StringVarP(&options.bind, "bind", "b", "", "Egress listener bind [PROTOCOL://[IP]:port]|[unix://socket]")

	return cmd
}

func (d *deleteCommand) validateOptions(options *deleteOptions) error {
	return nil
}

func (d *deleteCommand) run(cli cli.CLI, options *deleteOptions) error {
	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	req := graphql.DisableSidecarEgressInput{
		Selector: graphql.SidecarEgressSelector{
			Namespace: options.namespaceID,
		},
	}

	if options.workloadID != "" {
		workload, err := client.GetWorkloadWithSidecar(options.namespaceID, options.workloadID)
		if err != nil {
			return errors.WrapIf(err, "could not find workload in mesh, check the workload ID")
		}
		req.Selector.WorkloadLabels = &workload.Labels
	}

	if options.parsedBind != "" {
		req.Selector.Bind = &options.parsedBind
	}

	if options.parsedPort != nil {
		req.Selector.Port = options.parsedPort
	}

	response, err := client.DisableSidecarEgress(req)
	if err != nil {
		return errors.WrapIf(err, "could not delete sidecar egress")
	}

	if !response {
		return errors.New("unknown internal error: could not delete sidecar egress")
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

	log.Infof("sidecar egress for %s/%s deleted successfully\n\n", options.namespaceID, options.workloadID)

	return Output(cli, options.namespaceID, options.workloadID, sidecars, false, false)
}
