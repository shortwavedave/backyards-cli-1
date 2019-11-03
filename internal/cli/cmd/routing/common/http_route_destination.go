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
	"fmt"
	"strings"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

type HTTPRoute v1alpha3.HTTPRoute
type HTTPRouteDestination v1alpha3.HTTPRouteDestination
type HTTPRouteDestinations []HTTPRouteDestination

func (ds HTTPRouteDestinations) String() string {
	if len(ds) == 0 {
		return "-"
	}

	s := make([]string, len(ds))
	for k, v := range ds {
		s[k] = v.String()
	}
	return strings.Join(s, "\n")
}

func (d HTTPRouteDestination) String() string {
	var s string

	if d.Destination == nil {
		return s
	}

	s = Destination(*d.Destination).String()

	if d.Weight > 0 {
		s = fmt.Sprintf("%d%%%c%s", d.Weight, nbsp, s)
	}

	return s
}
