// Copyright © 2020 Banzai Cloud
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

package nodeexporter

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/banzaicloud/backyards-cli/pkg/helm"
)

type ChartValues struct {
	Image   helm.Image `json:"image,omitempty"`
	Service struct {
		Type        string            `json:"type,omitempty"`
		Port        int               `json:"port,omitempty"`
		TargetPort  int               `json:"targetPort,omitempty"`
		NodePort    int               `json:"nodePort,omitempty"`
		Annotations map[string]string `json:"annotations,omitempty"`
	} `json:"service,omitempty"`
	Resources      corev1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccount struct {
		Create bool   `json:"create,omitempty"`
		Name   string `json:"name,omitempty"`
	} `json:"serviceAccount,omitempty"`
	SecurityContext struct {
		RunASNonRoot bool `json:"runASNonRoot,omitempty"`
		RunAsUser    int  `json:"runAsUser,omitempty"`
	} `json:"securityContext,omitempty"`
	RBAC struct {
		Create     bool `json:"create,omitempty"`
		PSPEnabled bool `json:"pspEnabled,omitempty"`
	} `json:"rbac,omitempty"`
	HostNetwork bool                `json:"hostNetwork,omitempty"`
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	Prometheus  struct {
		Monitor struct {
			Enabled          bool              `json:"enabled,omitempty"`
			Namespace        string            `json:"namespace,omitempty"`
			AdditionalLabels map[string]string `json:"additionalLabels,omitempty"`
			ScrapeTimeout    string            `json:"scrapeTimeout,omitempty"`
		} `json:"monitor,omitempty"`
	} `json:"prometheus,omitempty"`
}

func (values *ChartValues) SetDefaults() {
	values.RBAC.PSPEnabled = true

	values.Service.Port = 19100
	values.Service.TargetPort = 19100
}
