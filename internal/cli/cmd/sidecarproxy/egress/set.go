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
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type setCommand struct{}

type setOptions struct {
	workloadID   string
	workloadName types.NamespacedName

	bind  string
	hosts []string

	parsedBind string
	parsedPort *v1alpha3.Port

	// TODO 2: configurable bind, port
}

func newSetCommand(cli cli.CLI) *cobra.Command {
	c := &setCommand{}
	options := &setOptions{}

	cmd := &cobra.Command{
		Use:           "set --workload namespace/[workload|*] [--bind [PROTOCOL://[IP]:port]|[unix://socket] [--hosts h1,h2]",
		Short:         "Set sidecar egress rule for a workload",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) > 0 {
				options.workloadID = args[0]
			}

			if options.workloadID == "" {
				return errors.New("workload must be specified")
			}

			options.workloadName, err = util.ParseK8sResourceIDAllowWildcard(options.workloadID)
			if err != nil {
				return errors.WrapIf(err, "could not parse workload ID")
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
	flags.StringVar(&options.workloadID, "workload", "", "Workload name [namespace/[workload|*]]")
	flags.StringVarP(&options.bind, "bind", "b", "", "Egress listener bind [PROTOCOL://[IP]:port]|[unix://socket]")
	flags.StringArrayVar(&options.hosts, "hosts", options.hosts, "Egress listener Hosts")

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

	req := graphql.ApplySidecarEgressInput{
		Selector: graphql.SidecarEgressSelector{
			Namespace: options.workloadName.Namespace,
		},
		Egress: graphql.Egress{
			Hosts: options.hosts,
		},
	}

	if options.workloadName.Name != "*" {
		workload, err := client.GetWorkloadSidecar(options.workloadName.Namespace, options.workloadName.Name)
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

	response, err := client.ApplySidecarEgress(req)
	if err != nil {
		return errors.WrapIf(err, "could not apply sidecar egress")
	}

	if !response {
		return errors.New("unknown internal error: could not apply sidecar egress")
	}

	var egressListeners map[string][]*v1alpha3.IstioEgressListener
	if options.workloadName.Name != "*" {
		workload, err := client.GetWorkloadSidecar(options.workloadName.Namespace, options.workloadName.Name)
		if err != nil {
			return err
		}
		egressListeners = common.GetEgressListenerMap(workload)
	} else {
		// TODO 1: implement and call GetNamespaceSidecars
		egressListeners = make(map[string][]*v1alpha3.IstioEgressListener)
	}

	log.Infof("sidecar egress for %s set successfully\n\n", options.workloadName)

	return Output(cli, options.workloadName, egressListeners)
}
