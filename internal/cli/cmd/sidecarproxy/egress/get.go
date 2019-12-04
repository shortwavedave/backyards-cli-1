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

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy/common"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type getCommand struct{}

type GetOptions struct {
	workloadID   string
	workloadName types.NamespacedName
}

func NewGetOptions() *GetOptions {
	return &GetOptions{}
}

func NewGetCommand(cli cli.CLI) *cobra.Command {
	c := &getCommand{}
	options := NewGetOptions()

	cmd := &cobra.Command{
		Use:           "get [[--workload=]namespace/workloadname]",
		Short:         "Get sidecar configuration for a workload",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) > 0 {
				options.workloadID = args[0]
			}

			if options.workloadID == "" {
				return errors.New("workload must be specified")
			}

			options.workloadName, err = util.ParseK8sResourceID(options.workloadID)
			if err != nil {
				return errors.WrapIf(err, "could not parse workload ID")
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.workloadID, "workload", "", "Workload name")

	return cmd
}

func (c *getCommand) run(cli cli.CLI, options *GetOptions) error {
	var err error

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	wl, err := client.GetWorkloadSidecar(options.workloadName.Namespace, options.workloadName.Name)
	if err != nil {
		return errors.WrapIf(err, "could not get workload")
	}

	if len(wl.Sidecars) == 0 {
		log.Infof("no sidecar found for %s", options.workloadName)
		return nil
	}

	egressRules := common.GetEgressListenerMap(wl)

	if len(egressRules) == 0 {
		log.Infof("no egress rule found for %s", options.workloadName)
		return nil
	}

	return Output(cli, options.workloadName, egressRules)
}
