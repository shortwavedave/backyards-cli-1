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

package route

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type getCommand struct{}

type GetOptions struct {
	serviceID string
	showAll   bool

	matches       []string
	parsedMatches []*v1alpha3.HTTPMatchRequest

	serviceName types.NamespacedName
}

func NewGetOptions() *GetOptions {
	return &GetOptions{
		showAll: true,
	}
}

func NewGetCommand(cli cli.CLI) *cobra.Command {
	c := &getCommand{}
	options := NewGetOptions()

	cmd := &cobra.Command{
		Use:           "get [[--service=]namespace/servicename] [[--match=]field:kind=value] ...",
		Short:         "Get route configuration for a service",
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

			options.serviceName, err = util.ParseK8sResourceID(options.serviceID)
			if err != nil {
				return errors.WrapIf(err, "could not parse service ID")
			}

			options.parsedMatches, err = common.ParseHTTPRequestMatches(options.matches)
			if err != nil {
				return errors.WrapIf(err, "could not parse matches")
			}

			if len(options.matches) > 0 {
				options.showAll = false
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.showAll, "show-all", "a", options.showAll, "Display settings for every route")
	flags.StringVar(&options.serviceID, "service", "", "Service name")
	flags.StringArrayVarP(&options.matches, "match", "m", options.matches, "HTTP request match")

	return cmd
}

func (c *getCommand) run(cli cli.CLI, options *GetOptions) error {
	var err error

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	service, err := client.GetService(options.serviceName.Namespace, options.serviceName.Name)
	if err != nil {
		return errors.WrapIf(err, "could not get service")
	}

	if len(service.VirtualServices) == 0 {
		log.Infof("no http route found for %s", options.serviceName)
		return nil
	}

	if options.showAll {
		return Output(cli, options.serviceName, service.VirtualServices[0].Spec.HTTP...)
	}

	matchedRoute := common.HTTPRoutes(service.VirtualServices[0].Spec.HTTP).GetMatchedRoute(options.parsedMatches)
	if matchedRoute == nil {
		return errors.Errorf("http route not found for %s", common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)))
	}

	return Output(cli, options.serviceName, *matchedRoute)
}
