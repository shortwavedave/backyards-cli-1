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

type HTTPRewrite v1alpha3.HTTPRewrite

func (r HTTPRewrite) String() string {
	var s string

	if r.Authority != nil && *r.Authority != "" {
		s = fmt.Sprintf("authority=%s", *r.Authority)
	}
	if r.URI != nil && *r.URI != "" {
		if s != "" {
			s += "\n"
		}
		s += fmt.Sprintf("uri=%s", *r.URI)
	}

	if s == "" {
		s = "-"
	}

	return s
}
