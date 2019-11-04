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
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/login"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/auth"

	"github.com/banzaicloud/backyards-cli/internal/platform/buildinfo"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

const (
	defaultVersionString = "unavailable"
	versionEndpoint      = "/version"
)

type versionCommand struct{}

type versionOptions struct {
	shortVersion      bool
	onlyClientVersion bool
}

func newVersionOptions() *versionOptions {
	return &versionOptions{
		shortVersion:      false,
		onlyClientVersion: false,
	}
}

func NewVersionCommand(cli cli.CLI) *cobra.Command {
	c := &versionCommand{}
	options := newVersionOptions()

	cmd := &cobra.Command{
		Use:           "version",
		Short:         "Print the client and api version information",
		Annotations:   map[string]string{util.CommandGroupAnnotationKey: util.HelperCommand},
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			c.run(cli, options)
		},
	}

	cmd.PersistentFlags().BoolVar(&options.shortVersion, "short", options.shortVersion, "Print the version number(s) only, with no additional output")
	cmd.PersistentFlags().BoolVar(&options.onlyClientVersion, "client", options.onlyClientVersion, "Print the client version only")

	return cmd
}

func (c *versionCommand) run(cli cli.CLI, options *versionOptions) {
	clientVersion := cli.GetRootCommand().Version
	if options.shortVersion {
		fmt.Println(clientVersion)
	} else {
		fmt.Fprintf(cli.Out(), "Client version: %s\n", clientVersion)
	}

	if options.onlyClientVersion {
		return
	}

	apiVersion := getAPIVersion(cli, versionEndpoint)
	if options.shortVersion {
		fmt.Println(apiVersion)
	} else {
		fmt.Fprintf(cli.Out(), "API version: %s\n", apiVersion)
	}
}

func getAPIVersion(cli cli.CLI, versionEndpoint string) string {
	endpoint, err := cli.InitializedEndpoint()
	if err != nil {
		logrus.Error(err)
		return defaultVersionString
	}
	defer endpoint.Close()

	token := cli.GetToken()
	if token == "" {
		err = login.Login(cli, func(authInfo *auth.Credentials) {
			token = authInfo.User.Token
			cli.GetPersistentConfig().SetToken(authInfo.User.Token)
		})
		if err != nil {
			logrus.Error(err)
			return defaultVersionString
		}
	}

	url := endpoint.URLForPath(versionEndpoint)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logrus.Error(err)
		return defaultVersionString
	}
	if token != "" {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// nolint G107
	resp, err := endpoint.HTTPClient().Do(request)
	if err != nil {
		logrus.Error(err)
		return defaultVersionString
	}
	defer resp.Body.Close()

	var bi buildinfo.BuildInfo
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&bi)
	if err != nil {
		logrus.Error(err)
		return defaultVersionString
	}

	if resp.StatusCode != 200 {
		logrus.Errorf("Request failed: %s", resp.Status)
		return defaultVersionString
	}

	if bi.Version != "" {
		return bi.Version
	}
	return defaultVersionString
}
