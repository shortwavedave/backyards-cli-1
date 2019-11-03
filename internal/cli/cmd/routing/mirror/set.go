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

package mirror

import (
	"regexp"
	"strconv"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/route"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/questionnaire"
)

type mirrorSettings struct {
	Host   string `json:"host" survey.question:"Destination host" survey.validate:"host"`
	Subset string `json:"subset" survey.question:"Destination subset" survey.validate:"subset"`
	Port   uint32 `json:"port,omitempty" survey.question:"Destination port" survey.validate:"port"`
}

type setCommand struct{}

type setOptions struct {
	serviceID string
	matches   []string
	mirrorSettings

	parsedMatches []*v1alpha3.HTTPMatchRequest
	serviceName   types.NamespacedName
}

func newSetOptions() *setOptions {
	return &setOptions{}
}

func newSetCommand(cli cli.CLI) *cobra.Command {
	c := &setCommand{}
	options := newSetOptions()

	cmd := &cobra.Command{
		Use:           "set [[--service=]namespace/servicename] [[--match=]field:kind=value] ...",
		Short:         "Set http route rewrite configuration of a service",
		Args:          cobra.ArbitraryArgs,
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

			if len(options.matches) == 0 {
				return errors.New("at least one route match must be specified")
			}

			options.parsedMatches, err = common.ParseHTTPRequestMatches(options.matches)
			if err != nil {
				return errors.WrapIf(err, "could not parse matches")
			}

			err = c.askQuestions(cli, options)
			if err != nil {
				return errors.WithStack(err)
			}

			err = c.validateOptions(options)
			if err != nil {
				return errors.WithStack(err)
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.serviceID, "service", "", "Service name")
	flags.StringArrayVarP(&options.matches, "match", "m", options.matches, "HTTP request match")

	flags.StringVar(&options.Host, "host", options.Host, "destination host")
	flags.StringVar(&options.Subset, "subset", options.Subset, "destination subset")
	flags.Uint32Var(&options.Port, "port", options.Port, "destination port")

	return cmd
}

func (c *setCommand) askQuestions(cli cli.CLI, options *setOptions) error {
	var err error

	if !cli.InteractiveTerminal() {
		return nil
	}

	qs, err := questionnaire.GetQuestionsFromStruct(options.mirrorSettings, map[string]survey.Validator{
		"host": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				if err := host(s).IsValid(); err != nil {
					return err
				}
			} else {
				return errors.New("invalid input type")
			}
			return nil
		},
		"subset": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				if err := subset(s).IsValid(); err != nil {
					return err
				}
			} else {
				return errors.New("invalid input type")
			}
			return nil
		},
		"port": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				if i, err := strconv.Atoi(s); err == nil {
					if err := port(i).IsValid(); err != nil {
						return err
					}
					return errors.WrapIf(err, "invalid port")
				}
			} else {
				return errors.New("invalid input type")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	err = survey.Ask(qs, &options.mirrorSettings)
	if err != nil {
		return errors.Wrap(err, "error while asking question")
	}

	return nil
}

func (c *setCommand) validateOptions(options *setOptions) error {
	if options.Host == "" {
		return errors.New("destination host must be specified")
	}

	if err := host(options.Host).IsValid(); err != nil {
		return err
	}

	if err := subset(options.Subset).IsValid(); err != nil {
		return err
	}

	if err := port(options.Port).IsValid(); err != nil {
		return err
	}

	return nil
}

func (c *setCommand) run(cli cli.CLI, options *setOptions) error {
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
		return errors.Errorf("http route not found for %s", common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)))
	}

	matchedRoute := common.HTTPRoutes(service.VirtualServices[0].Spec.HTTP).GetMatchedRoute(options.parsedMatches)
	if matchedRoute == nil {
		return errors.Errorf("http route not found for %s", common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)))
	}

	mirror := &v1alpha3.Destination{
		Host: options.Host,
	}
	if options.Subset != "" {
		mirror.Subset = &options.Subset
	}
	if options.Port > 0 {
		mirror.Port = &v1alpha3.PortSelector{
			Number: options.Port,
		}
	}

	req := graphql.ApplyHTTPRouteRequest{
		Selector: graphql.HTTPRouteSelector{
			Name:      service.Name,
			Namespace: service.Namespace,
			Matches:   options.parsedMatches,
		},
		Rule: graphql.HTTPRules{
			Mirror: mirror,
		},
	}

	r2, err := client.ApplyHTTPRoute(req)
	if err != nil {
		return err
	}

	if !r2 {
		return errors.New("unknown error: could not set mirror configuration")
	}

	log.Infof("mirror configuration for http route %s of %s set successfully", common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)), options.serviceName)

	service, err = client.GetService(options.serviceName.Namespace, options.serviceName.Name)
	if err != nil {
		return errors.WrapIf(err, "could not get service")
	}

	if len(service.VirtualServices) == 0 {
		return errors.Errorf("http route not found for %s", common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)))
	}

	matchedRoute = common.HTTPRoutes(service.VirtualServices[0].Spec.HTTP).GetMatchedRoute(options.parsedMatches)
	if matchedRoute == nil {
		return errors.Errorf("http route not found for %s", common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)))
	}

	return route.Output(cli, options.serviceName, *matchedRoute)
}

type port uint32

func (p port) IsValid() error {
	if p == 0 {
		return nil
	}

	if p > 65535 {
		return errors.Errorf("invalid port: %d", p)
	}

	return nil
}

type subset string

func (s subset) IsValid() error {
	if s == "" {
		return nil
	}

	authorityRegex := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
	if !authorityRegex.MatchString(string(s)) {
		return errors.Errorf("invalid subset: %s", s)
	}

	return nil
}

type host string

func (h host) IsValid() error {
	if h == "" {
		return errors.New("host must be specified")
	}

	uriRegex := regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])(\\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]))*$`)
	if !uriRegex.MatchString(string(h)) {
		return errors.Errorf("invalid host: %s", h)
	}

	return nil
}
