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

package graphql

import (
	"context"

	"github.com/MakeNowJust/heredoc"
)

type IntRange struct {
	Min uint `json:"min,omitempty"`
	Max uint `json:"max,omitempty"`
}

type GetAccessLogsInput struct {
	SourceType           string   `json:"sourceType,omitempty"`
	SourceName           string   `json:"sourceName,omitempty"`
	SourceNamespace      string   `json:"sourceNamespace,omitempty"`
	DestinationType      string   `json:"destinationType,omitempty"`
	DestinationName      string   `json:"destinationName,omitempty"`
	DestinationNamespace string   `json:"destinationNamespace,omitempty"`
	ReporterType         string   `json:"reporterType,omitempty"`
	ReporterName         string   `json:"reporterName,omitempty"`
	ReporterNamespace    string   `json:"reporterNamespace,omitempty"`
	Direction            string   `json:"direction,omitempty"`
	Authority            string   `json:"authority,omitempty"`
	Scheme               string   `json:"scheme,omitempty"`
	Method               string   `json:"method,omitempty"`
	Path                 string   `json:"path,omitempty"`
	StatusCode           IntRange `json:"statusCode,omitempty"`
}

func (c *client) SubscribeToAccessLogs(ctx context.Context, req *GetAccessLogsInput, resp chan interface{}, err chan error) {
	q := heredoc.Doc(`
	subscription getAccessLogs($input: AccessLogsInput) {
		accessLogs(input: $input) {
			reporter {
				id
				owner {
					raw
					type
					name
					namespace
				}
				clusterID
				instanceIPs
				istioVersion
				meshID
				name
				namespace
				workload
				serviceAccount
				metadata
			}
			direction
			startTime
			upstreamCluster
			source {
				address {
					ip
					port
				}
				name
				namespace
				workload
				serviceAccount
				metadata
			}
			destination {
				address {
					ip
					port
				}
				name
				namespace
				workload
				serviceAccount
				metadata
			}
			request {
				id
				method
				scheme
				authority
				path
				userAgent
				referer
				forwardedFor
				originalPath

				headerBytes
				bodyBytes
				metadata
				headers
			}
			response {
				statusCode
				statusCodeDetails
				flags

				headerBytes
				bodyBytes
				metadata
				headers
				trailers
			}
			latency: rawLatency
			durations: rawDurations {
				timeToLastRxByte
				timeToFirstUpstreamTxByte
				timeToLastUpstreamTxByte
				timeToFirstUpstreamRxByte
				timeToLastUpstreamRxByte
				timeToFirstDownstreamTxByte
				timeToLastDownstreamTxByte
			}
			protocolVersion
			authInfo {
				requestPrincipal
				principal
				namespace
				user
			}
		}
	}
	`)

	r := c.NewSubscribeRequest(q)
	r.Var("input", req)

	err <- c.WSClient().Subscribe(ctx, r, resp)
}
