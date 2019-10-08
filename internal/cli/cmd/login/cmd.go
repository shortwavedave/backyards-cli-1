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

package login

import (
	"github.com/banzaicloud/backyards-cli/pkg/servererror"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/pkg/auth"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

func NewLoginCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login",
		Aliases: []string{"l"},
		Short:   "Log in to Backyards",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := Login(cli, nil)
			return err
		},
	}

	return cmd
}

func Login(cli cli.CLI, onAuth func(*auth.ResponseBody)) (error) {
	authClient, err := common.GetAuthClient(cli)
	if err != nil {
		return err
	}
	authInfo, err := authClient.Login()
	if err != nil {
		if err != servererror.AuthDisabledError {
			return err
		}
	}
	if authInfo != nil {
		logrus.Infof("Logged in as %s", authInfo.User.Name)
		logrus.Debugf("Token: %s", authInfo.User.Token)
		logrus.Debugf("Wrapped token: %s", authInfo.User.WrappedToken)
		if onAuth != nil {
			onAuth(authInfo)
		}
	} else {
		logrus.Debug("Backyards authentication is disabled")
	}
	return nil
}
