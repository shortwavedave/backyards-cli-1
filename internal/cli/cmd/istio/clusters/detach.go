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

package clusters

import (
	"fmt"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type detachCommand struct {
	cli cli.CLI
}

type DetachOptions struct {
	name string
}

func NewDetachOptions() *DetachOptions {
	return &DetachOptions{}
}

func NewDetachCommand(cli cli.CLI, options *DetachOptions) *cobra.Command {
	c := &detachCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "detach [name]",
		Args:  cobra.ExactArgs(1),
		Short: "Detach peer cluster from the mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			options.name = args[0]

			return c.run(options)
		},
	}

	return cmd
}

func (c *detachCommand) run(options *DetachOptions) error {
	client, err := cmdCommon.GetGraphQLClient(c.cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	clusters, err := client.Clusters()
	if err != nil {
		return errors.WrapIf(err, "could not get clusters")
	}

	ok, peerCluster := clusters.GetClusterByName(options.name)
	if !ok || peerCluster.Type != "peer" {
		return errors.Errorf("peer cluster '%s' not found", options.name)
	}

	return c.cli.IfConfirmed(fmt.Sprintf("Detach peer cluster '%s'. Are you sure to proceed?", peerCluster.Name), func() error {
		ok, err = client.DetachPeerCluster(peerCluster.Name)
		if err != nil {
			return errors.WrapIf(err, "could not detach peer cluster")
		}

		if ok && c.cli.Interactive() {
			log.Infof("detaching peer cluster '%s' started successfully\n", options.name)
		}

		return nil
	})
}
