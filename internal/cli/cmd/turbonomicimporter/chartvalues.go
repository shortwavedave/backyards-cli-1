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

package turbonomicimporter

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/banzaicloud/backyards-cli/pkg/helm"
)

type Values struct {
	NameOverride         string                      `json:"nameOverride,omitempty"`
	Image                helm.Image                  `json:"image"`
	Resources            corev1.ResourceRequirements `json:"resources,omitempty"`
	UseNamespaceResource bool                        `json:"useNamespaceResource"`
	Turbonomic           struct {
		Hostname           string `json:"hostname,omitempty"`
		Username           string `json:"username,omitempty"`
		Password           string `json:"password,omitempty"`
		InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"`
	} `json:"turbonomic"`
}

func (values *Values) SetDefaults(hostname, username, password string, insecureSkipVerify bool) {
	values.UseNamespaceResource = true

	if hostname != "" {
		values.Turbonomic.Hostname = hostname
	}
	values.Turbonomic.Username = username
	values.Turbonomic.Password = password
	values.Turbonomic.InsecureSkipVerify = insecureSkipVerify
}
