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
	"io/ioutil"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

func NewEditCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit persistent configuration",
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		file := cli.GetPersistentConfig().GetConfigFileUsed()

		bytes, err := ioutil.ReadFile(cli.GetPersistentConfig().GetConfigFileUsed())
		if err != nil {
			logrus.Error(err, "failed to read config file")
		}

		content := string(bytes)
		err = survey.AskOne(&survey.Editor{
			Message:       file,
			Default:       content,
			HideDefault:   true,
			AppendDefault: true,
		}, &content)

		if err != nil {
			logrus.Error(err)
			return
		}

		fileInfo, err := os.Stat(file)
		if err != nil {
			logrus.Error(err)
		}

		err = ioutil.WriteFile(file, []byte(content), fileInfo.Mode())
		if err != nil {
			logrus.Error(err)
		}
		// short circuit here, to avoid overwriting manually edited contents
		os.Exit(0)
	}
	return cmd
}
