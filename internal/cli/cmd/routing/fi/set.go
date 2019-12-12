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
	"strconv"
	"time"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/questionnaire"
)

type FaultInjectionSettings struct {
	DelayPercentage float32 `json:"delayPercentage,omitempty" yaml:"delayPercentage,omitempty" survey.question:"Percentage of requests on which the delay will be injected"`
	DelayFixedDelay string  `json:"delayFixedDelay,omitempty" yaml:"delayFixedDelay,omitempty" survey.question:"Add a fixed delay before forwarding the request. Format: 1h/1m/1s/1ms. MUST be >1ms." survey.validate:"durationstring,fixed-delay"`

	AbortPercentage float32 `json:"abortPercentage,omitempty" yaml:"abortPercentage,omitempty" survey.question:"Percentage of requests on which the abort will be injected"`
	AborStatusCode  int     `json:"abortStatusCode,omitempty" yaml:"abortStatusCode,omitempty" survey.question:"HTTP status code to use to abort the HTTP request" survey.validate:"abort-status-code"`

	delayFixedDelay time.Duration
}

type setCommand struct{}

type setOptions struct {
	serviceID string
	matches   []string
	FaultInjectionSettings

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
		Short:         "Set fault injection for http route of a service",
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

			options.DelayFixedDelay = options.FaultInjectionSettings.delayFixedDelay.String()

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

	flags.Float32Var(&options.DelayPercentage, "delay-percentage", options.DelayPercentage, "Percentage of requests on which the delay will be injected")
	flags.DurationVar(&options.delayFixedDelay, "delay-fixed-delay", options.delayFixedDelay, "Add a fixed delay before forwarding the request. Format: 1h/1m/1s/1ms. MUST be >=1ms.")
	flags.Float32Var(&options.AbortPercentage, "abort-percentage", options.AbortPercentage, "Percentage of requests on which the abort will be injected")
	flags.IntVar(&options.AborStatusCode, "abort-status-code", options.AborStatusCode, "HTTP status code to use to abort the HTTP request")

	return cmd
}

func (c *setCommand) askQuestions(cli cli.CLI, options *setOptions) error {
	var err error

	if !cli.InteractiveTerminal() {
		return nil
	}

	qs, err := questionnaire.GetQuestionsFromStruct(options.FaultInjectionSettings, map[string]survey.Validator{
		"fixed-delay": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				if err := delayFixedDelay(s).IsValid(); err != nil {
					return err
				}
			} else {
				return errors.New("invalid input type")
			}
			return nil
		},
		"abort-status-code": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				n, err := strconv.Atoi(s)
				if err != nil {
					return err
				}
				err = abortStatusCode(n).IsValid()
				if err != nil {
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

	err = survey.Ask(qs, &options.FaultInjectionSettings)
	if err != nil {
		return errors.Wrap(err, "error while asking question")
	}

	return nil
}

func (c *setCommand) validateOptions(options *setOptions) error {
	if err := delayFixedDelay(options.DelayFixedDelay).IsValid(); err != nil {
		return err
	}

	if err := abortStatusCode(options.AborStatusCode).IsValid(); err != nil {
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

	fi := &v1alpha3.HTTPFaultInjection{
		Delay: &v1alpha3.Delay{
			FixedDelay: options.DelayFixedDelay,
			Percentage: &v1alpha3.Percentage{
				Value: options.DelayPercentage,
			},
		},
		Abort: &v1alpha3.Abort{
			HTTPStatus: options.AborStatusCode,
			Percentage: &v1alpha3.Percentage{
				Value: options.AbortPercentage,
			},
		},
	}

	req := graphql.ApplyHTTPRouteRequest{
		Selector: graphql.HTTPRouteSelector{
			Name:      service.Name,
			Namespace: service.Namespace,
			Matches:   options.parsedMatches,
		},
		Rule: graphql.HTTPRules{
			FaultInjection: fi,
		},
	}

	r2, err := client.ApplyHTTPRoute(req)
	if err != nil {
		return err
	}

	if !r2 {
		return errors.New("unknown error: could not set fault injection")
	}

	log.Infof("fault injection for %s set successfully", options.serviceName)

	return Output(cli, options.serviceName, v1alpha3.HTTPRoute{
		Match: options.parsedMatches,
		Fault: fi,
	})
}

type delayFixedDelay string

func (fd delayFixedDelay) IsValid() error {
	if d, err := time.ParseDuration(string(fd)); err == nil {
		if d.Milliseconds() <= 1 {
			return errors.New("fixed delay must be greater than 1ms")
		}
	} else {
		return errors.WrapIf(err, "could not parse duration")
	}

	return nil
}

type abortStatusCode int

func (c abortStatusCode) IsValid() error {
	if c < 400 || c > 599 {
		return errors.New("abort status code must be in 400-599 range")
	}

	return nil
}
