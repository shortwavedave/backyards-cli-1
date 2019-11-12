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
	"bufio"
	"fmt"
	"os"
	"sync"

	"emperror.dev/errors"
	"github.com/MakeNowJust/heredoc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/auth"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/servererror"
)

var mutex sync.Mutex
var inMemoryAuthInfo *auth.Credentials

func NewLoginCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "login",
		Aliases:     []string{"l"},
		Short:       "Log in to Backyards",
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.OperationCommand},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if cli.InteractiveTerminal() {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print(heredoc.Doc(`
					The following token will be valid for a few seconds to log in over the UI.

					Notes:
					 - use the "dashboard" command to open a browser tab and log in automatically
					 - rerun this command in case you need a fresh token

					Press enter to continue.
				`))
				_, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
			}
			err = Login(cli, func(body *auth.Credentials) {
				if cli.InteractiveTerminal() {
					logrus.Infof("Login token: %s", body.User.WrappedToken)
					cli.GetPersistentConfig().SetToken(body.User.Token)
				} else {
					fmt.Println(body.User.WrappedToken)
				}
			})
			return err
		},
	}

	return cmd
}

func Login(cli cli.CLI, onAuth func(*auth.Credentials)) error {
	mutex.Lock()
	defer mutex.Unlock()
	if inMemoryAuthInfo != nil {
		if onAuth != nil {
			onAuth(inMemoryAuthInfo)
		}
		return nil
	}
	config, err := cli.GetK8sConfig()
	if err != nil {
		return err
	}

	endpoint, err := cli.InitializedEndpoint()
	if err != nil {
		return err
	}
	defer endpoint.Close()

	url := endpoint.URLForPath("/api/login")
	if endpoint.CA() != nil {
		config.TLSClientConfig.CAData = endpoint.CA()
	}

	authClient := auth.NewClient(config, url)
	authInfo, err := authClient.Login()
	if err != nil {
		if err != servererror.ErrAuthDisabled {
			return errors.Wrap(err, "failed to log in, you may need to install backyards first")
		}
	}
	if authInfo != nil {
		inMemoryAuthInfo = authInfo
		if cli.InteractiveTerminal() {
			logrus.Infof("Logged in as %s", authInfo.User.Name)
			logrus.Debugf("Token: %s", authInfo.User.Token)
			logrus.Debugf("Wrapped token: %s", authInfo.User.WrappedToken)
		}
		cli.GetPersistentConfig().SetToken(authInfo.User.Token)
		if onAuth != nil {
			onAuth(authInfo)
		}
	} else if cli.InteractiveTerminal() {
		logrus.Debug("Backyards authentication is disabled")
	}
	return nil
}
