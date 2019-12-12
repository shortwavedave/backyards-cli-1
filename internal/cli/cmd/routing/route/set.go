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
	"strings"
	"time"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type RedirectSettings struct {
	URI       string `json:"uri"`
	Authority string `json:"authority"`
}

type RetriesSettings struct {
	Attempts      int    `json:"attempts"`
	PerTryTimeout string `json:"perTryTimeout,omitempty"`
	RetryOn       string `json:"retryOn,omitempty"`

	perTryTimeout time.Duration
}

type TimeoutSettings struct {
	Timeout string `json:"timeout,omitempty"`

	timeout time.Duration
}

type setCommand struct{}

type setOptions struct {
	serviceID    string
	matches      []string
	destinations []string
	weights      string

	RedirectSettings
	RetriesSettings
	TimeoutSettings

	parsedMatches      []*v1alpha3.HTTPMatchRequest
	parsedDestinations []v1alpha3.Destination
	parsedWeights      []int
	serviceName        types.NamespacedName
}

func newSetOptions() *setOptions {
	return &setOptions{
		TimeoutSettings: TimeoutSettings{
			timeout: -1,
		},
		RetriesSettings: RetriesSettings{
			Attempts:      -1,
			perTryTimeout: time.Second * 2,
		},
	}
}

func newSetCommand(cli cli.CLI) *cobra.Command {
	c := &setCommand{}
	options := newSetOptions()

	cmd := &cobra.Command{
		Use:           "set [[--service=]namespace/servicename] [[--match=]field:kind=value] ...",
		Short:         "Set http route for a service",
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

			options.parsedMatches, err = common.ParseHTTPRequestMatches(options.matches)
			if err != nil {
				return errors.WrapIf(err, "could not parse matches")
			}

			options.parsedDestinations, err = common.ParseDestinations(options.destinations)
			if err != nil {
				return errors.WrapIf(err, "could not parse destinations")
			}

			options.parsedWeights, err = common.ParseWeights(options.weights)
			if err != nil {
				return errors.WrapIf(err, "could not parse weights")
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

	flags.StringArrayVarP(&options.destinations, "route-destination", "d", options.destinations, "HTTP route destination")
	flags.StringVarP(&options.weights, "route-weights", "w", "100", "The proportions of traffic to be forwarded to the route destinations. (0-100)")

	flags.StringVar(&options.URI, "redirect-uri", options.URI, "overwrite the Path portion of the URL with this value")
	flags.StringVar(&options.Authority, "redirect-authority", options.Authority, "overwrite the Authority/Host portion of the URL with this value")

	flags.DurationVarP(&options.timeout, "timeout", "t", options.timeout, "Timeout for HTTP requests")

	flags.StringVar(&options.RetryOn, "retry-on", options.RetryOn, "Specifies the conditions under which retry takes place")
	flags.IntVar(&options.Attempts, "retry-attempts", options.Attempts, "Number of retries for a given request")
	flags.DurationVar(&options.perTryTimeout, "retry-per-try-timeout", options.perTryTimeout, "Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms, must be >=1ms")

	return cmd
}

func (c *setCommand) validateOptions(options *setOptions) error {
	if len(options.matches) == 0 {
		return errors.New("at least one route match must be specified")
	}

	if len(options.destinations) > 0 && (options.RedirectSettings.URI != "" || options.RedirectSettings.Authority != "") {
		return errors.New("http route cannot contain both route destination and redirect")
	}

	if len(options.parsedDestinations) > 0 && len(options.parsedDestinations) != len(options.parsedWeights) {
		return errors.New("weight must be set for all route destinations")
	}

	if options.RetryOn != "" {
		rons := make([]string, 0)
		for _, ron := range strings.Split(options.RetryOn, ",") {
			if !common.SupportedRetryOnPolicies[ron] {
				return errors.Errorf("invalid retry-on policy: %s", ron)
			}
			ron = strings.TrimSpace(ron)
			rons = append(rons, ron)
		}
		options.RetryOn = strings.Join(rons, ",")
	}

	if options.perTryTimeout.Milliseconds() < 2 {
		return errors.New("per-try-timeout must be > 1ms")
	}

	return nil
}

func (c *setCommand) run(cli cli.CLI, options *setOptions) error {
	disableRules := make([]string, 0)

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	service, err := client.GetService(options.serviceName.Namespace, options.serviceName.Name)
	if err != nil {
		return errors.WrapIf(err, "could not get service")
	}

	routes := make([]*v1alpha3.HTTPRouteDestination, 0)
	for i, d := range options.parsedDestinations {
		d := d
		routes = append(routes, &v1alpha3.HTTPRouteDestination{
			Destination: &d,
			Weight:      &options.parsedWeights[i],
		})
	}

	var redirect *v1alpha3.HTTPRedirect
	if options.RedirectSettings.URI != "" {
		redirect = &v1alpha3.HTTPRedirect{}
		redirect.URI = &options.RedirectSettings.URI
	}
	if options.RedirectSettings.Authority != "" {
		if redirect == nil {
			redirect = &v1alpha3.HTTPRedirect{}
		}
		redirect.Authority = &options.RedirectSettings.Authority
	}

	var retries *v1alpha3.HTTPRetry
	if options.Attempts > 0 {
		retries = &v1alpha3.HTTPRetry{}
		retries.Attempts = options.Attempts
	} else {
		options.perTryTimeout = 0
	}
	if options.perTryTimeout.Milliseconds() > 1 {
		if retries == nil {
			retries = &v1alpha3.HTTPRetry{}
		}
		retries.PerTryTimeout = options.perTryTimeout.String()
	}
	if options.RetryOn != "" && options.Attempts > 0 {
		if retries == nil {
			retries = &v1alpha3.HTTPRetry{}
		}
		retries.RetryOn = &options.RetryOn
	}
	if retries == nil && options.Attempts == 0 {
		disableRules = append(disableRules, "Retries")
	}

	var timeout *string
	if options.timeout > 0 {
		t := options.timeout.String()
		timeout = &t
	}

	if options.timeout == 0 {
		disableRules = append(disableRules, "Timeout")
	}

	if len(disableRules) > 0 {
		_, err := client.DisableHTTPRoute(graphql.DisableHTTPRouteRequest{
			Selector: graphql.HTTPRouteSelector{
				Name:      service.Name,
				Namespace: service.Namespace,
				Matches:   options.parsedMatches,
			},
			Rules: disableRules,
		})
		if err != nil {
			return errors.WrapIff(err, "could not disable rules: %s", strings.Join(disableRules, ","))
		}
	}

	if len(routes) == 0 && redirect == nil && timeout == nil && retries == nil && len(disableRules) == 0 {
		log.Info("routes/redirect, timeout or retry settings must be specified")
		return nil
	}

	req := graphql.ApplyHTTPRouteRequest{
		Selector: graphql.HTTPRouteSelector{
			Name:      service.Name,
			Namespace: service.Namespace,
			Matches:   options.parsedMatches,
		},
		Rule: graphql.HTTPRules{
			Matches:  options.parsedMatches,
			Route:    routes,
			Redirect: redirect,
			Timeout:  timeout,
			Retries:  retries,
		},
	}

	response, err := client.ApplyHTTPRoute(req)
	if err != nil {
		return errors.WrapIf(err, "could not apply http route")
	}

	if !response {
		return errors.New("unknown error: could not set http route")
	}

	service, err = client.GetService(service.Namespace, service.Name)
	if err != nil {
		return err
	}

	log.Infof("routing for %s set successfully\n\n", options.serviceName)

	matchedRoute := common.HTTPRoutes(service.VirtualServices[0].Spec.HTTP).GetMatchedRoute(options.parsedMatches)
	if matchedRoute == nil {
		return errors.New("invalid response")
	}

	return Output(cli, options.serviceName, *matchedRoute)
}
