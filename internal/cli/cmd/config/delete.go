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
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

func NewDeleteCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete persistent configuration",
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		file := cli.GetPersistentConfig().GetConfigFileUsed()

		confirm := false
		err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Are you sure to delete %s?", file),
			Default: true,
		}, &confirm)
		if err != nil {
			logrus.Error(err)
			return
		}

		err = os.Remove(file)
		if err != nil {
			logrus.Error(err)
		}
		// short circuit here, to avoid inconsistent state
		os.Exit(0)
	}
	return cmd
}
