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

package routing

import (
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/cb"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/fi"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/route"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/ts"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

func NewRootCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "routing",
		Aliases:     []string{"r"},
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.OperationCommand},
		Short:       "Manage service routing configurations",
	}

	cmd.AddCommand(
		ts.NewRootCmd(cli),
		cb.NewRootCmd(cli),
		fi.NewRootCmd(cli),
		route.NewRootCmd(cli),
	)

	return cmd
}
