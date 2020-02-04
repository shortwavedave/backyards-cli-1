// Copyright © 2020 Banzai Cloud
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

package istio

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type outboundTrafficCommand struct {
	cli cli.CLI
}

func NewOutboundTrafficCommand(cli cli.CLI) *cobra.Command {
	c := &outboundTrafficCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:     "outbound-traffic-policy [allowed|restricted]",
		Aliases: []string{"otp"},
		Args:    cobra.MaximumNArgs(1),
		Short:   "Set outbound traffic policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if len(args) == 1 {
				var mode string
				switch args[0] {
				case "allowed":
					mode = "ALLOW_ANY"
				case "restricted":
					mode = "REGISTRY_ONLY"
				default:
					return errors.Errorf("invalid outbound mesh policy: %s", args[0])
				}
				return c.set(cli, mode)
			}

			return c.show(cli)
		},
	}

	return cmd
}

func (c *outboundTrafficCommand) show(cli cli.CLI) error {
	client, err := cli.GetK8sClient()
	if err != nil {
		return errors.WrapIf(err, "could not get k8s client")
	}

	istio, err := FetchIstioCR(client)
	if err != nil {
		return errors.WithStackIf(err)
	}

	_, err = cli.Out().Write([]byte(fmt.Sprintf("mesh wide outbound traffic policy is currently set to '%s'\n", istio.Spec.OutboundTrafficPolicy.Mode)))
	if err != nil {
		return errors.WithStackIf(err)
	}

	return nil
}

func (c *outboundTrafficCommand) set(cli cli.CLI, mode string) error {
	client, err := cli.GetK8sClient()
	if err != nil {
		return errors.WrapIf(err, "could not get k8s client")
	}

	istio, err := FetchIstioCR(client)
	if err != nil {
		return errors.WithStackIf(err)
	}

	if istio.Spec.OutboundTrafficPolicy.Mode == mode {
		_, err = cli.Out().Write([]byte(fmt.Sprintf("mesh wide outbound traffic policy is already set to '%s'\n", istio.Spec.OutboundTrafficPolicy.Mode)))
		if err != nil {
			return errors.WithStackIf(err)
		}
		return nil
	}

	istio.Spec.OutboundTrafficPolicy.Mode = mode

	err = client.Update(context.Background(), istio)
	if err != nil {
		return errors.WrapIf(err, "could not set outbound traffic policy")
	}

	_, err = cli.Out().Write([]byte(fmt.Sprintf("mesh wide outbound traffic policy is set to '%s'\n", istio.Spec.OutboundTrafficPolicy.Mode)))
	if err != nil {
		return errors.WithStackIf(err)
	}

	return nil
}
