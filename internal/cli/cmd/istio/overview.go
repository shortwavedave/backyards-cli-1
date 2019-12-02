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

package istio

import (
	"fmt"

	"emperror.dev/errors"
	"github.com/spf13/cobra"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/output"
)

const defaultEvaluationDurationSeconds = 60

type overviewCommand struct {
	cli cli.CLI
}

type OverviewOptions struct {
	evaluationDurationSeconds uint
}

func NewOverviewOptions(evaluationDurationSeconds uint) *OverviewOptions {
	return &OverviewOptions{
		evaluationDurationSeconds: evaluationDurationSeconds,
	}
}

func NewOverviewCommand(cli cli.CLI, options *OverviewOptions) *cobra.Command {
	c := &overviewCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "overview",
		Args:  cobra.NoArgs,
		Short: "Show basic mesh overview metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if options.evaluationDurationSeconds < 1 {
				return errors.New("evaluationDurationSeconds must be greater than 0")
			}

			return c.run(cli, options)
		},
	}

	cmd.Flags().UintVarP(&options.evaluationDurationSeconds, "evaluation-duration-seconds", "s", options.evaluationDurationSeconds, "Metrics timespan in seconds")

	return cmd
}

func (c *overviewCommand) run(cli cli.CLI, options *OverviewOptions) error {
	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	a, err := client.Overview(options.evaluationDurationSeconds)
	if err != nil {
		return err
	}

	if cli.OutputFormat() == output.OutputFormatTable && cli.Interactive() {
		fmt.Fprintf(cli.Out(), "Mesh overview – metrics time span %d seconds\n\n", options.evaluationDurationSeconds)
	}

	err = c.output(cli, a)
	if err != nil {
		return err
	}

	return nil
}

func (c *overviewCommand) output(cli output.FormatContext, data interface{}) error {
	ctx := &output.Context{
		Out:     cli.Out(),
		Color:   cli.Color(),
		Format:  cli.OutputFormat(),
		Fields:  []string{"Clusters", "Services", "ServicesInMesh", "Workloads", "WorkloadsInMesh", "Pods", "PodsInMesh", "ErrorRate", "Latency", "RPS"},
		Headers: []string{"Clusters", "Services", "in mesh", "Workloads", "in mesh", "Pods", "in mesh", "Error rate", "Latency", "RPS"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		return errors.WrapIf(err, "could not produce output")
	}

	return nil
}
