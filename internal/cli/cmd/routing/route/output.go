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
	"fmt"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/output"
)

type Out struct {
	Matches  common.HTTPMatchRequests     `json:"matches,omitempty"`
	Routes   common.HTTPRouteDestinations `json:"routes,omitempty"`
	Redirect common.HTTPRedirect          `json:"redirect,omitempty"`
	Timeout  common.Timeout               `json:"timeout,omitempty"`
	Retries  common.HTTPRetry             `json:"retries,omitempty"`
	Rewrite  common.HTTPRewrite           `json:"rewrite,omitempty"`
	Mirror   common.Destination           `json:"mirror,omitempty"`
}

func Output(cli cli.CLI, serviceName types.NamespacedName, routes ...v1alpha3.HTTPRoute) error {
	var err error

	outs := make([]Out, 0)
	for _, route := range routes {
		o := Out{}
		o.Matches = common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(route.Match))
		routes := make(common.HTTPRouteDestinations, len(route.Route))
		for k, v := range route.Route {
			if v == nil {
				continue
			}
			routes[k] = common.HTTPRouteDestination(*v)
		}
		o.Routes = routes
		if route.Redirect != nil {
			r := common.HTTPRedirect(*route.Redirect)
			o.Redirect = r
		}
		if route.Timeout != nil {
			o.Timeout = common.Timeout(*route.Timeout)
		}
		if route.Retries != nil {
			o.Retries = common.HTTPRetry(*route.Retries)
		}
		if route.Rewrite != nil {
			o.Rewrite = common.HTTPRewrite(*route.Rewrite)
		}
		if route.Mirror != nil {
			o.Mirror = common.Destination(*route.Mirror)
		}

		outs = append(outs, o)
	}

	if cli.OutputFormat() == output.OutputFormatTable && cli.Interactive() {
		fmt.Fprintf(cli.Out(), "Settings for %s\n\n", serviceName)
	}

	err = show(cli, outs)
	if err != nil {
		return err
	}

	if cli.Interactive() {
		fmt.Println()
	}

	return nil
}

func show(cli output.FormatContext, data interface{}) error {
	ctx := &output.Context{
		Out:     cli.Out(),
		Color:   cli.Color(),
		Format:  cli.OutputFormat(),
		Fields:  []string{"Matches", "Routes", "Redirect", "Timeout", "Retries", "Rewrite", "Mirror"},
		Headers: []string{"Matches", "Routes", "Redirect", "Timeout", "Retry", "Rewrite", "Mirror To"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		return errors.WrapIf(err, "could not produce output")
	}

	return nil
}
