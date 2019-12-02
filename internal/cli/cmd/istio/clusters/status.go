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
	"strings"

	"emperror.dev/errors"
	"github.com/spf13/cobra"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/output"
)

type statusCommand struct {
	cli cli.CLI
}

func NewStatusCommand(cli cli.CLI) *cobra.Command {
	c := &statusCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.NoArgs,
		Short: "Show cluster status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(cli)
		},
	}

	return cmd
}

func (c *statusCommand) run(cli cli.CLI) error {
	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	clusters, err := client.Clusters()
	if err != nil {
		return err
	}

	type Clusters struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		Namespace      string   `json:"namespace"`
		Type           string   `json:"type"`
		Status         string   `json:"status"`
		ErrorMessage   string   `json:"errorMessage"`
		GatewayAddress []string `json:"gatewayAddress"`
	}

	data := make([]Clusters, len(clusters))
	for i, c := range clusters {
		data[i] = Clusters{
			ID:             c.ID,
			Name:           c.Name,
			Namespace:      c.Namespace,
			Type:           strings.Title(c.Type),
			Status:         c.Status.Status,
			ErrorMessage:   c.Status.ErrorMessage,
			GatewayAddress: c.Status.GatewayAddress,
		}
	}

	if cli.OutputFormat() == output.OutputFormatTable && cli.Interactive() {
		fmt.Fprintf(cli.Out(), "Clusters in the mesh\n\n")
	}

	err = c.output(cli, data)
	if err != nil {
		return err
	}

	return nil
}

func (c *statusCommand) output(cli output.FormatContext, data interface{}) error {
	ctx := &output.Context{
		Out:     cli.Out(),
		Color:   cli.Color(),
		Format:  cli.OutputFormat(),
		Fields:  []string{"Name", "Type", "Status", "GatewayAddress", "ErrorMessage"},
		Headers: []string{"Name", "Type", "Status", "Gateway Address", "Message"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		return errors.WrapIf(err, "could not produce output")
	}

	return nil
}
