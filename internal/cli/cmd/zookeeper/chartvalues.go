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

package zookeeper

import (
	"github.com/banzaicloud/backyards-cli/pkg/helm"
)

type Values struct {
	NameOverride         string     `json:"nameOverride,omitempty"`
	FullnameOverride     string     `json:"fullnameOverride,omitempty"`
	ReplicaCount         int        `json:"replicaCount"`
	UseNamespaceResource bool       `json:"useNamespaceResource"`
	Image                helm.Image `json:"image"`

	RBAC struct {
		Enabled    bool   `json:"enabled"`
		APIVersion string `json:"apiVersion"`
	} `json:"rbac"`
}