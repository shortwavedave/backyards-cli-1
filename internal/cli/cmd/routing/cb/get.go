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

package cb

import (
	"fmt"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	clierrors "github.com/banzaicloud/backyards-cli/internal/errors"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/output"
)

type getCommand struct{}

type getOptions struct {
	serviceID string

	serviceName types.NamespacedName
}

func newGetOptions() *getOptions {
	return &getOptions{}
}

func newGetCommand(cli cli.CLI) *cobra.Command {
	c := &getCommand{}
	options := newGetOptions()

	cmd := &cobra.Command{
		Use:           "get [[--service=]namespace/servicename]",
		Short:         "Get circuit breaker rules for a service",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) > 0 {
				options.serviceID = args[0]
			}

			if options.serviceID == "" {
				return errors.New("service must be specified")
			}

			options.serviceName, err = common.ParseServiceID(options.serviceID)
			if err != nil {
				return err
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.serviceID, "service", "", "Service name")

	return cmd
}

func getCircuitBreakerRulesByServiceName(cli cli.CLI, serviceName types.NamespacedName) (*CircuitBreakerSettings, error) {
	var err error

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	service, err := client.GetService(serviceName.Namespace, serviceName.Name)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get service")
	}

	if len(service.DestinationRules) == 0 {
		return nil, clierrors.NotFoundError{}
	}

	drule := service.DestinationRules[0]
	if drule.Spec.TrafficPolicy == nil || drule.Spec.TrafficPolicy.ConnectionPool == nil {
		return nil, clierrors.NotFoundError{}
	}

	tp := drule.Spec.TrafficPolicy
	cb := CircuitBreakerSettings{}

	if tp.ConnectionPool != nil {
		if tp.ConnectionPool.TCP != nil {
			if tp.ConnectionPool.TCP.MaxConnections != nil {
				cb.MaxConnections = *tp.ConnectionPool.TCP.MaxConnections
			}
			if tp.ConnectionPool.TCP.ConnectTimeout != nil {
				cb.ConnectTimeout = *tp.ConnectionPool.TCP.ConnectTimeout
			}
		}
		if tp.ConnectionPool.HTTP != nil && tp.ConnectionPool.HTTP.HTTP1MaxPendingRequests != nil {
			cb.HTTP1MaxPendingRequests = *tp.ConnectionPool.HTTP.HTTP1MaxPendingRequests
		}
		if tp.ConnectionPool.HTTP != nil && tp.ConnectionPool.HTTP.HTTP2MaxRequests != nil {
			cb.HTTP2MaxRequests = *tp.ConnectionPool.HTTP.HTTP2MaxRequests
		}
		if tp.ConnectionPool.HTTP != nil && tp.ConnectionPool.HTTP.MaxRequestsPerConnection != nil {
			cb.MaxRequestsPerConnection = *tp.ConnectionPool.HTTP.MaxRequestsPerConnection
		}
		if tp.ConnectionPool.HTTP != nil && tp.ConnectionPool.HTTP.MaxRetries != nil {
			cb.MaxRetries = *tp.ConnectionPool.HTTP.MaxRetries
		}
	}

	if tp.OutlierDetection != nil {
		cb.ConsecutiveErrors = tp.OutlierDetection.ConsecutiveErrors
		if tp.OutlierDetection.Interval != nil {
			cb.Interval = *tp.OutlierDetection.Interval
		}
		if tp.OutlierDetection.BaseEjectionTime != nil {
			cb.BaseEjectionTime = *tp.OutlierDetection.BaseEjectionTime
		}
		if tp.OutlierDetection.MaxEjectionPercent != nil {
			cb.MaxEjectionPercent = *tp.OutlierDetection.MaxEjectionPercent
		}
	}

	return &cb, nil
}

func (c *getCommand) run(cli cli.CLI, options *getOptions) error {
	var err error

	data, err := getCircuitBreakerRulesByServiceName(cli, options.serviceName)
	if err != nil {
		if clierrors.IsNotFound(err) {
			log.Infof("no circuit breaker rules set for %s", options.serviceName)
			return nil
		}
		return err
	}

	if cli.OutputFormat() == output.OutputFormatTable && cli.Interactive() {
		fmt.Fprintf(cli.Out(), "Settings for %s\n\n", options.serviceName)
	}

	err = Output(cli, data)
	if err != nil {
		return err
	}

	return nil
}
