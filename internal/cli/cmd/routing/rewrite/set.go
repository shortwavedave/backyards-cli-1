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

package rewrite

import (
	"regexp"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/route"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/questionnaire"
)

type rewriteSettings struct {
	Authority string `json:"authority,omitempty" survey.question:"Authority/Host header will be rewritten with this value" survey.validate:"authority"`
	URI       string `json:"uri,omitempty" survey.question:"Path (or the prefix) portion of the URI will be rewritten with this value" survey.validate:"uri"`
}

type setCommand struct{}

type setOptions struct {
	serviceID string
	matches   []string
	rewriteSettings

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

			options.serviceName, err = util.ParseK8sResourceID(options.serviceID)
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

	flags.StringVarP(&options.Authority, "authority", "a", options.Authority, "Path (or the prefix) portion of the URI will be rewritten with this value.")
	flags.StringVarP(&options.URI, "uri", "u", options.URI, "Authority/Host header will be rewritten with this value")

	return cmd
}

func (c *setCommand) askQuestions(cli cli.CLI, options *setOptions) error {
	var err error

	if !cli.InteractiveTerminal() {
		return nil
	}

	qs, err := questionnaire.GetQuestionsFromStruct(options.rewriteSettings, map[string]survey.Validator{
		"authority": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				if err := authority(s).IsValid(); err != nil {
					return err
				}
			} else {
				return errors.New("invalid input type")
			}
			return nil
		},
		"uri": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				if err := uri(s).IsValid(); err != nil {
					return err
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

	err = survey.Ask(qs, &options.rewriteSettings)
	if err != nil {
		return errors.Wrap(err, "error while asking question")
	}

	return nil
}

func (c *setCommand) validateOptions(options *setOptions) error {
	if options.Authority == "" && options.URI == "" {
		return errors.New("either URI, Authority or both must be specified")
	}

	if err := authority(options.Authority).IsValid(); err != nil {
		return err
	}

	if err := uri(options.URI).IsValid(); err != nil {
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

	rr := &v1alpha3.HTTPRewrite{
		URI:       &options.URI,
		Authority: &options.Authority,
	}

	req := graphql.ApplyHTTPRouteRequest{
		Selector: graphql.HTTPRouteSelector{
			Name:      service.Name,
			Namespace: service.Namespace,
			Matches:   options.parsedMatches,
		},
		Rule: graphql.HTTPRules{
			Rewrite: rr,
		},
	}

	r2, err := client.ApplyHTTPRoute(req)
	if err != nil {
		return err
	}

	if !r2 {
		return errors.New("unknown error: could not set rewrite configuration")
	}

	log.Infof("rewrite configuration for http route %s of %s set successfully", common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(options.parsedMatches)), options.serviceName)

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

type authority string

func (a authority) IsValid() error {
	if a == "" {
		return nil
	}

	authorityRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(:(\\d{1,5}))?$`)
	if !authorityRegex.MatchString(string(a)) {
		return errors.Errorf("invalid authority: %s", a)
	}

	return nil
}

type uri string

func (u uri) IsValid() error {
	if u == "" {
		return nil
	}

	uriRegex := regexp.MustCompile("^[a-zA-Z0-9/._~!$&'()*+,;=:@%?/]+$")
	if !uriRegex.MatchString(string(u)) {
		return errors.Errorf("invalid uri: %s", u)
	}

	return nil
}
