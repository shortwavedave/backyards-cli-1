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
	"time"

	al_proto "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

type HTTPAccessLogEntry struct {
	Reporter        *Reporter         `json:"reporter"`
	Direction       string            `json:"direction"`
	StartTime       string            `json:"startTime"`
	UpstreamCluster string            `json:"upstreamCluster"`
	Source          *RequestEndpoint  `json:"source"`
	Destination     *RequestEndpoint  `json:"destination"`
	Request         *HTTPRequest      `json:"request"`
	Response        *HTTPResponse     `json:"response"`
	Latency         *time.Duration    `json:"latency"`
	Durations       *RequestDurations `json:"durations"`
	ProtocolVersion string            `json:"protocolVersion"`
	AuthInfo        *AuthInfo         `json:"authInfo,omitempty"`

	startTime *time.Time
	entry     *al_proto.HTTPAccessLogEntry
}

type AuthInfo struct {
	RequestPrincipal string `json:"requestPrincipal,omitempty"`
	Principal        string `json:"principal,omitempty"`
	Namespace        string `json:"namespace,omitempty"`
	User             string `json:"user,omitempty"`
}

type Reporter struct {
	ID             string            `json:"id,omitempty"`
	Owner          *Owner            `json:"owner,omitempty"`
	ClusterID      string            `json:"clusterID,omitempty"`
	InstanceIPs    []string          `json:"instanceIPs,omitempty"`
	IstioVersion   string            `json:"istioVersion,omitempty"`
	MeshID         string            `json:"meshID,omitempty"`
	Name           string            `json:"name,omitempty"`
	Namespace      string            `json:"namespace,omitempty"`
	Workload       string            `json:"workload,omitempty"`
	ServiceAccount string            `json:"serviceAccount,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type Owner struct {
	Raw       string `json:"raw,omitempty"`
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type RequestDurations struct {
	// time it takes to receive a request
	TimeToLastRxByte *time.Duration `json:"timeToLastRxByte,omitempty"`
	// time it takes to send out a request from the proxy
	TimeToFirstUpstreamTxByte *time.Duration `json:"timeToFirstUpstreamTxByte,omitempty"`
	// time it takes for a request to be sent from the proxy
	TimeToLastUpstreamTxByte *time.Duration `json:"timeToLastUpstreamTxByte,omitempty"`
	// time it takes to start receiving a response
	TimeToFirstUpstreamRxByte *time.Duration `json:"timeToFirstUpstreamRxByte,omitempty"`
	// time it takes to receive a complete response
	TimeToLastUpstreamRxByte *time.Duration `json:"timeToLastUpstreamRxByte,omitempty"`
	// time it takes to start sending a response to downstream
	TimeToFirstDownstreamTxByte *time.Duration `json:"timeToFirstDownstreamTxByte,omitempty"`
	// time it takes to send a complete response to downstream
	TimeToLastDownstreamTxByte *time.Duration `json:"timeToLastDownstreamTxByte,omitempty"`
}

type HTTPRequest struct {
	ID           string `json:"id"`
	Method       string `json:"method"`
	Scheme       string `json:"scheme"`
	Authority    string `json:"authority"`
	Path         string `json:"path"`
	UserAgent    string `json:"userAgent,omitempty"`
	Referer      string `json:"referer,omitempty"`
	ForwardedFor string `json:"forwardedFor,omitempty"`
	OriginalPath string `json:"originalPath,omitempty"`

	HeaderBytes uint64            `json:"headerBytes"`
	BodyBytes   uint64            `json:"bodyBytes"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type HTTPResponse struct {
	StatusCode        uint32   `json:"statusCode"`
	StatusCodeDetails string   `json:"statusCodeDetails"`
	Flags             []string `json:"flags"`

	HeaderBytes uint64            `json:"headerBytes"`
	BodyBytes   uint64            `json:"bodyBytes"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Trailers    map[string]string `json:"trailers,omitempty"`
}

type RequestEndpoint struct {
	Address        *TCPAddr          `json:"address"`
	Name           string            `json:"name,omitempty"`
	Namespace      string            `json:"namespace,omitempty"`
	Workload       string            `json:"workload,omitempty"`
	ServiceAccount string            `json:"serviceAccount,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type TCPAddr struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
