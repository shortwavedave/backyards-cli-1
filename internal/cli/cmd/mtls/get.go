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

type getCommand struct{}

func NewGetCommand(cli cli.CLI) *cobra.Command {
	c := &getCommand{}
	options := newMTLSOptions()

	cmd := &cobra.Command{
		Use:           "get [[--resource=]mesh|namespace|namespace/servicename]",
		Short:         "Get mTLS policy setting for a resource",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := parseMTLSArgs(options, args, cli, true, false)
			if err != nil {
				return errors.WrapIf(err, "could not parse arguments")
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.resourceID, "resource", "", "Resource name")

	return cmd
}

func (c *getCommand) run(cli cli.CLI, options *mTLSOptions) error {
	var err error

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	switch {
	case options.resourceName.Name == meshWidePolicy:
		return getMesh(cli, options, client)
	case options.resourceName.Name == "":
		return getNamespace(cli, options, client)
	default:
		return getService(cli, options, client)
	}
}
