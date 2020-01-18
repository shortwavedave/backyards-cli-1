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

package ale

import (
	"regexp"
	"strings"
)

var ownerMatchRegexp = regexp.MustCompile(`^kubernetes://.*/namespaces/([a-z0-9]+(?:-[a-z0-9]+)*)/([a-z0-9]+(?:-[a-z0-9]+)*)/([a-z0-9]+(?:-[a-z0-9]+)*)$`)

func (re *Reporter) IsIdentifiable() {}

func (re *Reporter) SetAttributes(attrs map[string]interface{}) {
	if value, ok := attrs["NAME"].(string); ok {
		re.Name = value
	}
	if value, ok := attrs["NAMESPACE"].(string); ok {
		re.Namespace = value
	}
	if value, ok := attrs["WORKLOAD_NAME"].(string); ok {
		re.Workload = value
	}
	if value, ok := attrs["SERVICE_ACCOUNT"].(string); ok {
		re.ServiceAccount = value
	}
	if value, ok := attrs["CLUSTER_ID"].(string); ok {
		re.ClusterID = value
	}
	if value, ok := attrs["INSTANCE_IPS"].(string); ok {
		re.InstanceIPs = strings.Split(value, ",")
	}
	if value, ok := attrs["OWNER"].(string); ok {
		m := ownerMatchRegexp.FindStringSubmatch(value)
		if len(m) == 4 {
			re.Owner = &Owner{
				Raw:       value,
				Type:      m[2],
				Name:      m[3],
				Namespace: m[1],
			}
		}
	}
	if value, ok := attrs["ISTIO_VERSION"].(string); ok {
		re.IstioVersion = value
	}
	if value, ok := attrs["MESH_ID"].(string); ok {
		re.MeshID = value
	}

	if re.Metadata == nil {
		re.Metadata = make(map[string]string)
	}

	if labels, ok := attrs["LABELS"].(map[string]interface{}); ok {
		for k, v := range labels {
			if value, ok := v.(string); ok {
				re.Metadata[k] = value
			}
		}
	}

	for k, v := range attrs {
		if value, ok := v.(string); ok {
			re.Metadata[k] = value
		}
	}
}
