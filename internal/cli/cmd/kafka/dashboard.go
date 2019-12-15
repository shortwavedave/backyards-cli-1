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
	"os"
	"os/signal"

	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type dashboardCommand struct{}

func NewDashboardCommand(cli cli.CLI) *cobra.Command {
	c := dashboardCommand{}

	cmd := &cobra.Command{
		Use:         "dashboard",
		Short:       "Open the Kafka Grafana dashboard in a web browser",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.OperationCommand},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.run(cli)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}

func (c *dashboardCommand) run(cli cli.CLI) error {
	var err error
	var url string

	endpoint, err := cli.InitializedEndpoint()
	if err != nil {
		return err
	}

	url = endpoint.URLForPath("/grafana/d/r2jsS3-Zk/envoy-kafka-protocol-filter?orgId=1&var-brokerId=0")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer func() {
		<-signals
		signal.Stop(signals)
	}()

	log.Infof("Opening Kafka Grafana dashboard at %s", url)

	err = browser.OpenURL(url)
	if err != nil {
		return err
	}

	return nil
}
