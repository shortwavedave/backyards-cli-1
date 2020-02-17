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

package mtls

import (
	"emperror.dev/errors"
	"github.com/spf13/cobra"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type disableCommand struct{}

func NewDisableCommand(cli cli.CLI) *cobra.Command {
	c := &disableCommand{}
	options := newMTLSOptions()

	cmd := &cobra.Command{
		Use:           "disable [[--resource=]mesh|namespace|namespace/servicename[:[portname|portnumber]]]",
		Short:         "Set mTLS policy setting for a resource to DISABLED",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := parseMTLSArgs(options, args, cli, true, true)
			if err != nil {
				return errors.WrapIf(err, "could not parse arguments")
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.resourceID, "resource", "", "Resource name")
	flags.StringVar(&options.portName, "portName", "", "Port name")
	flags.IntVar(&options.portNumber, "portNumber", 0, "Port number")

	return cmd
}

func (c *disableCommand) run(cli cli.CLI, options *mTLSOptions) error {
	var err error

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	return setMTLS(cli, options, client, ModeDisabled)
}
