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
	"context"

	"github.com/machinebox/graphql"

	"github.com/banzaicloud/backyards-cli/internal/endpoint"
)

type Client interface {
	WSClient() *WSClient
	NewSubscribeRequest(q string) *Request
	SetJWTToken(string)
	GetNamespaces() (NamespacesResponse, error)
	GetPod(namespace, name string) (*Pod, error)
	GetNamespace(name string) (NamespaceResponse, error)
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
	GetWorkload(namespace, name string) (*MeshWorkload, error)
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
	ApplyMeshPolicy(input ApplyMeshPolicyInput) (bool, error)
	SubscribeToAccessLogs(ctx context.Context, req *GetAccessLogsInput, resp chan interface{}, err chan error)
	Close()
}

type client struct {
	jwtToken string
	endpoint endpoint.Endpoint
	client   *graphql.Client
	wsClient *WSClient
}

func NewClient(endpoint endpoint.Endpoint, path string) Client {
	url := endpoint.URLForPath(path)
	return &client{
		client:   graphql.NewClient(url, graphql.WithHTTPClient(endpoint.HTTPClient())),
		wsClient: NewWSClient(url, WithHTTPClient(endpoint.HTTPClient())),
		endpoint: endpoint,
	}
}

func (c *client) SetJWTToken(token string) {
	c.jwtToken = token
}

func (c *client) NewRequest(q string) *graphql.Request {
	r := c.newRequest(q)
	gr := graphql.NewRequest(q)
	gr.Header = r.GetHeader()

	return gr
}

func (c *client) NewSubscribeRequest(q string) *Request {
	return c.newRequest(q)
}

func (c *client) newRequest(q string) *Request {
	r := &Request{}

	// set query string
	r.Query(q)

	// set header fields
	if c.jwtToken != "" {
		r.GetHeader().Set("Authorization", "Bearer "+c.jwtToken)
	}
	r.GetHeader().Set("Cache-Control", "no-cache")

	return r
}

func (c *client) WSClient() *WSClient {
	return c.wsClient
}

func (c *client) Close() {
	c.endpoint.Close()
}
