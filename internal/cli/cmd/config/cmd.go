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

package config

import (
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

func NewConfigCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "View and manage persistent configuration",
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.HelperCommand},
	}

	cmd.AddCommand(
		NewViewCmd(cli),
		NewEditCmd(cli),
		NewDeleteCmd(cli),
	)

	return cmd
}
