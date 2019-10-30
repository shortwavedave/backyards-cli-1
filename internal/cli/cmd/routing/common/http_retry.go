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

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

// SupportedRetryOnPolicies contains envoy supported retry on header values
var SupportedRetryOnPolicies = map[string]bool{
	// 'x-envoy-retry-on' supported policies:
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-on
	"5xx":                    true,
	"gateway-error":          true,
	"connect-failure":        true,
	"retriable-4xx":          true,
	"refused-stream":         true,
	"retriable-status-codes": true,

	// 'x-envoy-retry-grpc-on' supported policies:
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-grpc-on
	"cancelled":          true,
	"deadline-exceeded":  true,
	"internal":           true,
	"resource-exhausted": true,
	"unavailable":        true,
}

type HTTPRetry v1alpha3.HTTPRetry

func (r HTTPRetry) String() string {
	var s string
	if r.Attempts > 0 {
		s = fmt.Sprintf("%dx", r.Attempts)
	}

	if r.PerTryTimeout != "" {
		s = fmt.Sprintf("%s (%s ptt)", s, r.PerTryTimeout)
	}

	if r.RetryOn != nil {
		s = fmt.Sprintf("%s on %s", s, *r.RetryOn)
	}

	if s != "" {
		return s
	}

	return "-"
}
