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

package common

import (
	"reflect"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

type HTTPRoutes []v1alpha3.HTTPRoute

func (r HTTPRoutes) GetMatchedRoute(matches []*v1alpha3.HTTPMatchRequest) *v1alpha3.HTTPRoute {
	var matchedRoute *v1alpha3.HTTPRoute

	for _, route := range r {
		route := route
		if len(matches) == 0 && len(route.Match) == 0 {
			matchedRoute = &route
			break
		}
		if reflect.DeepEqual(matches, route.Match) {
			matchedRoute = &route
			break
		}
	}

	return matchedRoute
}
