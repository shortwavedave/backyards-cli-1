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

func (re *RequestEndpoint) String() string {
	if re.Name != "" {
		return re.Name
	}

	if re.Metadata["authority"] != "" {
		return re.Metadata["authority"]
	}

	return re.Address.IP
}

func (re *RequestEndpoint) SetAttributes(attrs map[string]interface{}) {
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

	if value, ok := attrs["OWNER"].(string); ok {
		re.Metadata["owner"] = value
	}
}
