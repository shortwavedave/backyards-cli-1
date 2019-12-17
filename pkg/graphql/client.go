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
	"github.com/machinebox/graphql"

	"github.com/banzaicloud/backyards-cli/internal/endpoint"
)

type Client interface {
	SetJWTToken(string)
	GetNamespaces() (NamespacesResponse, error)
	GetNamespaceWithSidecar(name string) (NamespaceResponse, error)
	GetNamespaceWithSidecarRecommendation(name string, isolationLevel string) (NamespaceResponse, error)
	GetNamespaceWithMTLS(name string) (NamespaceResponse, error)
	EnableAutoSidecarInjection(req EnableAutoSidecarInjectionRequest) (EnableAutoSidecarInjectionResponse, error)
	DisableAutoSidecarInjection(req DisableAutoSidecarInjectionRequest) (DisableAutoSidecarInjectionResponse, error)
	GenerateLoad(req GenerateLoadRequest) (GenerateLoadResponse, error)
	ApplyHTTPRoute(req ApplyHTTPRouteRequest) (ApplyHTTPRouteResponse, error)
	DisableHTTPRoute(req DisableHTTPRouteRequest) (DisableHTTPRouteResponse, error)
	ApplyGlobalTrafficPolicy(req ApplyGlobalTrafficPolicyRequest) (ApplyGlobalTrafficPolicyResponse, error)
	DisableGlobalTrafficPolicy(req DisableGlobalTrafficPolicyRequest) (DisableGlobalTrafficPolicyResponse, error)
	GetService(namespace, name string) (*MeshService, error)
	GetWorkloadWithSidecar(namespace, name string) (*MeshWorkloadSidecar, error)
	GetWorkloadWithSidecarRecommendation(namespace, name string, isolationLevel string, labelWhitelist []string) (*MeshWorkloadSidecar, error)
	GetServiceWithMTLS(namespace, name string) (*MeshService, error)
	Overview(evaluationDurationSeconds uint) (OverviewResponse, error)
	GetMeshWithMTLS() (*MeshPolicy, error)
	Clusters() (ClustersResponse, error)
	AttachPeerCluster(req AttachPeerClusterRequest) (bool, error)
	DetachPeerCluster(name string) (bool, error)
	ApplySidecarEgress(input ApplySidecarEgressInput) (ApplySidecarEgressResponse, error)
	DisableSidecarEgress(input DisableSidecarEgressInput) (DisableSidecarEgressResponse, error)
	ApplyPolicyPeers(input ApplyPolicyPeersInput) (bool, error)
	DisablePolicyPeers(input DisablePolicyPeersInput) (bool, error)
	SwitchGlobalMTLS(enabled bool) (bool, error)
	Close()
}

type client struct {
	jwtToken string
	endpoint endpoint.Endpoint
	client   *graphql.Client
}

func NewClient(endpoint endpoint.Endpoint, path string) Client {
	return &client{
		client:   graphql.NewClient(endpoint.URLForPath(path), graphql.WithHTTPClient(endpoint.HTTPClient())),
		endpoint: endpoint,
	}
}

func (c *client) SetJWTToken(token string) {
	c.jwtToken = token
}

func (c *client) NewRequest(q string) *graphql.Request {
	r := graphql.NewRequest(q)

	// set header fields
	if c.jwtToken != "" {
		r.Header.Set("Authorization", "Bearer "+c.jwtToken)
	}
	r.Header.Set("Cache-Control", "no-cache")

	return r
}

func (c *client) Close() {
	c.endpoint.Close()
}
