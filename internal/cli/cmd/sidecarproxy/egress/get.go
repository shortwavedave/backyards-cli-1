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

	"github.com/banzaicloud/backyards-cli/pkg/graphql"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type getCommand struct{}

type GetOptions struct {
	workloadID  string
	namespaceID string
}

func NewGetOptions() *GetOptions {
	return &GetOptions{}
}

func NewGetCommand(cli cli.CLI) *cobra.Command {
	c := &getCommand{}
	options := NewGetOptions()

	cmd := &cobra.Command{
		Use:           "get [--namespace] namespace [--workload name]",
		Short:         "Get sidecar configuration for a namespace or a workload",
		Args:          cobra.MaximumNArgs(2),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.namespaceID, "namespace", "", "Namespace name")
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

	sidecars, err := getSidecars(client, options.namespaceID, options.workloadID)
	if err != nil {
		return errors.WrapIf(err, "could not retrieve sidecars through graphql")
	}

	return Output(cli, options.namespaceID, options.workloadID, sidecars, false, false)
}

func getSidecars(client graphql.Client, namespace, name string) ([]graphql.Sidecar, error) {
	var sidecars []graphql.Sidecar
	if name != "" {
		wl, err := client.GetWorkloadWithSidecar(namespace, name)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't query workload sidecars")
		}

		if len(wl.Sidecars) == 0 {
			log.Infof("no sidecar found for %s/%s", namespace, name)
			return nil, nil
		}

		sidecars = wl.Sidecars
	} else {
		resp, err := client.GetNamespaceWithSidecar(namespace)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't query namespace sidecars")
		}
		sidecars = resp.Namespace.Sidecars
	}
	return sidecars, nil
}
