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

package graphql

import (
	"context"

	"github.com/MakeNowJust/heredoc"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

type ApplyHTTPRouteRequest struct {
	Selector HTTPRouteSelector `json:"selector"`
	Rule     HTTPRules         `json:"rule"`
}

type HTTPRules struct {
	Matches        []*v1alpha3.HTTPMatchRequest     `json:"match,omitempty"`
	Route          []*v1alpha3.HTTPRouteDestination `json:"route,omitempty"`
	Redirect       *v1alpha3.HTTPRedirect           `json:"redirect,omitempty"`
	FaultInjection *v1alpha3.HTTPFaultInjection     `json:"fault,omitempty"`
	Timeout        *string                          `json:"timeout,omitempty"`
	Retries        *v1alpha3.HTTPRetry              `json:"retries,omitempty"`
	Rewrite        *v1alpha3.HTTPRewrite            `json:"rewrite,omitempty"`
	Mirror         *v1alpha3.Destination            `json:"mirror,omitempty"`
}

type HTTPRouteSelector struct {
	Name      string                       `json:"name"`
	Namespace string                       `json:"namespace"`
	Matches   []*v1alpha3.HTTPMatchRequest `json:"match,omitempty"`
}

type ApplyHTTPRouteResponse bool

func (c *client) ApplyHTTPRoute(req ApplyHTTPRouteRequest) (ApplyHTTPRouteResponse, error) {
	request := heredoc.Doc(`
	  mutation applyHTTPRoute(
		$input: ApplyHTTPRouteInput!
	  ) {
		applyHTTPRoute(
		  input: $input
		)
	  }
`)

	r := c.NewRequest(request)
	r.Var("input", req)

	// run it and capture the response
	var respData map[string]ApplyHTTPRouteResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return false, err
	}

	return respData["applyHTTPRoute"], nil
}
