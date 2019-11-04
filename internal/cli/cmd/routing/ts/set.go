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

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type setCommand struct{}

type setOptions struct {
	serviceID string
	matches   []string
	subsets   []string

	serviceName   types.NamespacedName
	parsedSubsets parsedSubsets
	parsedMatches []*v1alpha3.HTTPMatchRequest
}

func newSetOptions() *setOptions {
	return &setOptions{}
}

func newSetCommand(cli cli.CLI) *cobra.Command {
	c := &setCommand{}
	options := newSetOptions()

	cmd := &cobra.Command{
		Use:           "set [[--service=]namespace/servicename] [[--match=]field:kind=value] ... [[--version=]subset=weight] ...",
		Short:         "Set traffic shifting rules for a service",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) > 0 {
				options.serviceID = args[0]
			}

			if len(args) > 1 {
				options.subsets = append(options.subsets, args[1:]...)
			}

			if options.serviceID == "" {
				return errors.New("service must be specified")
			}

			if len(options.matches) == 0 {
				return errors.New("at least one route match must be specified")
			}

			if len(options.subsets) < 1 {
				return errors.New("at least 1 subset must be specified")
			}

			options.serviceName, err = common.ParseServiceID(options.serviceID)
			if err != nil {
				return err
			}

			options.parsedMatches, err = common.ParseHTTPRequestMatches(options.matches)
			if err != nil {
				return errors.WrapIf(err, "could not parse matches")
			}

			options.parsedSubsets, err = parseSubsets(options.subsets)
			if err != nil {
				return err
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.serviceID, "service", "", "Service name")
	flags.StringArrayVarP(&options.subsets, "subset", "s", []string{}, "Subsets with weights (sum of the weight must add up to 100)")
	flags.StringArrayVarP(&options.matches, "match", "m", options.matches, "HTTP request match")

	return cmd
}

func (c *setCommand) run(cli cli.CLI, options *setOptions) error {
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

	req := graphql.ApplyHTTPRouteRequest{
		Selector: graphql.HTTPRouteSelector{
			Name:      service.Name,
			Namespace: service.Namespace,
			Matches:   options.parsedMatches,
		},
		Rule: graphql.HTTPRules{
			Matches: options.parsedMatches,
			Route:   make([]*v1alpha3.HTTPRouteDestination, 0),
		},
	}

	for subset, weight := range options.parsedSubsets {
		subset := subset
		req.Rule.Route = append(req.Rule.Route, &v1alpha3.HTTPRouteDestination{
			Destination: &v1alpha3.Destination{
				Host:   service.Name,
				Subset: &subset,
			},
			Weight: weight,
		})
	}

	r, err := client.ApplyHTTPRoute(req)
	if err != nil {
		return err
	}

	if !r {
		return errors.New("unknown error: cannot set traffic shifting")
	}

	log.Infof("traffic shifting for %s with match %s set to %s successfully", options.serviceName, common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)), options.parsedSubsets)

	return nil
}
