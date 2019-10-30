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

package fi

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type getCommand struct{}

type getOptions struct {
	serviceID string
	showAll   bool

	matches       []string
	parsedMatches []*v1alpha3.HTTPMatchRequest

	serviceName types.NamespacedName
}

func newGetOptions() *getOptions {
	return &getOptions{
		showAll: true,
	}
}

func newGetCommand(cli cli.CLI) *cobra.Command {
	c := &getCommand{}
	options := newGetOptions()

	cmd := &cobra.Command{
		Use:           "get [[--service=]namespace/servicename] [[--match=]field:kind=value] ...",
		Short:         "Get fault injection rules for a service",
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

func (c *getCommand) run(cli cli.CLI, options *getOptions) error {
	var err error

	client, err := common.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	service, err := client.GetService(options.serviceName.Namespace, options.serviceName.Name)
	if err != nil {
		return errors.WrapIf(err, "could not get service")
	}

	r, err := client.GetService(service.Namespace, service.Name)
	if err != nil {
		return errors.WrapIf(err, "could not get service")
	}

	if len(r.VirtualServices) == 0 {
		log.Infof("No fault injection settings found for %s", options.serviceName)
		return nil
	}

	if options.showAll {
		return Output(cli, options.serviceName, r.VirtualServices[0].Spec.HTTP...)
	}

	matchedRoute := common.HTTPRoutes(r.VirtualServices[0].Spec.HTTP).GetMatchedRoute(options.parsedMatches)

	if matchedRoute == nil {
		return errors.New("route not found")
	}

	return Output(cli, options.serviceName, *matchedRoute)
}
