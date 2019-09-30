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
	"fmt"
	neturl "net/url"
	"os"
	"os/signal"

	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/login"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type dashboardCommand struct{}

type DashboardOptions struct {
	QueryParams  map[string]string
	Login        bool
	WrappedToken string
}

func NewDashboardOptions() *DashboardOptions {
	return &DashboardOptions{
		QueryParams: make(map[string]string),
	}
}

func NewDashboardCommand(cli cli.CLI, options *DashboardOptions) *cobra.Command {
	c := dashboardCommand{}

	cmd := &cobra.Command{
		Use:   "dashboard [flags]",
		Short: "Open the Backyards dashboard in a web browser",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if options.Login {
				authInfo, err := login.Login(cli)
				if err != nil {
					return err
				}
				options.WrappedToken = authInfo.User.WrappedToken
			}
			err = c.run(cli, options)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&options.Login, "login", options.Login,
		"Login to Backyards automatically using Kubernetes credentials")

	return cmd
}

func (c *dashboardCommand) run(cli cli.CLI, options *DashboardOptions) error {
	var err error
	var url string

	url, err = cli.GetEndpointURL("")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer func() {
		<-signals
		signal.Stop(signals)
	}()

	if err != nil {
		return err
	}

	url, err = withQueryParams(url, options.QueryParams)
	if err != nil {
		return err
	}

	log.Infof("Opening Backyards UI at %s", url)

	if options.WrappedToken != "" {
		url, err = withQueryParams(url, map[string]string{"wrapped-token": options.WrappedToken})
		if err != nil {
			return err
		}
		url, err = withPath(url, "/api/login")
		if err != nil {
			return err
		}
		log.Debugf("Open %s", url)
	}

	err = browser.OpenURL(url)
	if err != nil {
		return err
	}

	return nil
}

func withQueryParams(url string, params map[string]string) (string, error) {
	uri, err := neturl.ParseRequestURI(url)
	if err != nil {
		return "", err
	}

	q := uri.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	uri.RawQuery = q.Encode()

	return uri.String(), nil
}

func withPath(url string, path string) (string, error) {
	uri, err := neturl.ParseRequestURI(url)
	if err != nil {
		return "", err
	}

	uri.Path = fmt.Sprintf("%s%s", uri.Path, path)

	return uri.String(), nil
}
