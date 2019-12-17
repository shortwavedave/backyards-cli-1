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
	"errors"

	"github.com/MakeNowJust/heredoc"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

type Sidecar struct {
	v1alpha3.Sidecar
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type MeshWorkloadSidecar struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`

	Sidecars            []Sidecar `json:"sidecars"`
	RecommendedSidecars []Sidecar `json:"recommendedSidecars"`
}

func (c *client) GetWorkloadWithSidecar(namespace, name string) (*MeshWorkloadSidecar, error) {
	request := heredoc.Doc(`
	query($namespace: String!, $name: String!) {
      workload(namespace: $namespace, name: $name) {
        id
        name
        namespace
        labels
        sidecars {
          name
          namespace
          spec {
            workloadSelector {
              labels
            }
            egress {
              port {
                number
                protocol
                name
              }
              bind
              captureMode
              hosts
            }
            ingress {
              port {
                number
                protocol
                name
              }
              bind
              captureMode
              defaultEndpoint
            }
            outboundTrafficPolicy
          }
        }
      }
    }
`)

	type Response struct {
		Workload MeshWorkloadSidecar `json:"workload"`
	}

	r := c.NewRequest(request)
	r.Var("name", name)
	r.Var("namespace", namespace)

	// run it and capture the response
	var respData Response
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return nil, err
	}

	if respData.Workload.ID == "" {
		return nil, errors.New("not found")
	}

	return &respData.Workload, nil
}

func (c *client) GetWorkloadWithSidecarRecommendation(namespace, name string, isolationLevel string, labelWhitelist []string) (*MeshWorkloadSidecar, error) {
	request := heredoc.Doc(`
	query($namespace: String!, $name: String!, $isolationLevel: IsolationLevel, $labelWhitelist: [String!]) {
      workload(namespace: $namespace, name: $name) {
        id
        name
        namespace
        labels
        recommendedSidecars(isolationLevel: $isolationLevel, labelWhitelist: $labelWhitelist) {
          name
          namespace
          spec {
            workloadSelector {
              labels
            }
            egress {
              port {
                number
                protocol
                name
              }
              bind
              captureMode
              hosts
            }
          }
        }
      }
    }
`)

	type Response struct {
		Workload MeshWorkloadSidecar `json:"workload"`
	}

	r := c.NewRequest(request)
	r.Var("name", name)
	r.Var("namespace", namespace)
	if isolationLevel != "" {
		r.Var("isolationLevel", isolationLevel)
	}
	r.Var("labelWhitelist", labelWhitelist)

	// run it and capture the response
	var respData Response
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return nil, err
	}

	if respData.Workload.ID == "" {
		return nil, errors.New("not found")
	}

	return &respData.Workload, nil
}
