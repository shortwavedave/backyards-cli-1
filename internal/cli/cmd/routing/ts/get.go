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

package ts

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/route"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
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
		Short:         "Get traffic shifting rules for a service",
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
				return err
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
	flags.StringVar(&options.serviceID, "service", "", "Service name")
	flags.StringArrayVarP(&options.matches, "match", "m", options.matches, "HTTP request match")

	return cmd
}

func (c *getCommand) run(cli cli.CLI, options *getOptions) error {
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
		log.Infof("no routing configuration found for %s", options.serviceName)
		return nil
	}

	if options.showAll {
		return route.Output(cli, options.serviceName, service.VirtualServices[0].Spec.HTTP...)
	}

	matchedRoute := common.HTTPRoutes(service.VirtualServices[0].Spec.HTTP).GetMatchedRoute(options.parsedMatches)
	if matchedRoute == nil {
		log.Infof("no route found for %s", options.serviceName)
	}

	return route.Output(cli, options.serviceName, *matchedRoute)
}
