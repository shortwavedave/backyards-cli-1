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

package autoinject

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type offCommand struct{}

type offOptions struct {
	namespaceName string
}

func newOffOptions() *offOptions {
	return &offOptions{}
}

func newOffCommand(cli cli.CLI) *cobra.Command {
	c := &offCommand{}
	options := newOffOptions()

	cmd := &cobra.Command{
		Use:           "off [[--namespace=]name]",
		Short:         "Disable sidecar injection for the given namespace",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) > 0 {
				options.namespaceName = args[0]
			}

			if options.namespaceName == "" {
				if cli.Interactive() {
					var err error
					options.namespaceName, err = common.GetNamespaceNamesInteractively(cli)
					if err != nil {
						return err
					}
				} else {
					return errors.New("namespace name must be specified")
				}
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.namespaceName, "namespace", "", "Namespace name")

	return cmd
}

func (c *offCommand) run(cli cli.CLI, options *offOptions) error {
	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get graphql client")
	}
	req := graphql.DisableAutoSidecarInjectionRequest{
		Name: options.namespaceName,
	}
	defer client.Close()

	resp, err := client.DisableAutoSidecarInjection(req)
	if err != nil {
		return err
	}
	if len(resp.NameSpaces) == 0 {
		return errors.New("unknown error occurred")
	}

	log.Infof("auto sidecar injection successfully removed from namespace %s", options.namespaceName)

	return nil
}
