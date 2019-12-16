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

package kafka

import (
	"context"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type loadCommand struct {
	cli cli.CLI
}

type LoadOptions struct {
	replicas int32
}

func NewLoadOptions() *LoadOptions {
	return &LoadOptions{}
}

func NewLoadCommand(cli cli.CLI, options *LoadOptions) *cobra.Command {
	c := &loadCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "load [flags]",
		Args:  cobra.NoArgs,
		Short: "Load kafka with perf tool",
		Long: `Loads kafka with perf tool.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.`,
		Example: `  # Default install.
  backyards kafka load`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(options)
		},
	}

	cmd.Flags().Int32Var(&options.replicas, "replicas", 2, "How many replicas the perf tool should create")

	return cmd
}

func (c *loadCommand) run(options *LoadOptions) error {
	client, err := c.cli.GetK8sClient()
	if err != nil {
		return err
	}
	perfDeployment := appsv1.Deployment{}
	err = client.Get(context.Background(), types.NamespacedName{Name: "perf-load", Namespace: kafkaNamespace}, &perfDeployment)
	if err != nil {
		return err
	}
	perfDeployment.Spec.Replicas = &options.replicas

	err = client.Update(context.Background(), &perfDeployment)
	if err != nil {
		return err
	}

	return nil
}
