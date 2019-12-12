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

	"github.com/MakeNowJust/heredoc"
)

type IstioNamespace struct {
	Name     string    `json:"name"`
	Sidecars []Sidecar `json:"sidecars"`
}

type NamespacesResponse struct {
	Namespaces []IstioNamespace `json:"namespaces"`
}

type NamespaceResponse struct {
	Namespace IstioNamespace `json:"namespace"`
}

func (c *client) GetNamespaces() (NamespacesResponse, error) {
	request := heredoc.Doc(`
		query namespaces {
			namespaces{
				name
			}
		}`)
	r := c.NewRequest(request)
	var respData NamespacesResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return respData, err
	}

	return respData, nil
}

func (c *client) GetNamespaceWithSidecar(name string) (NamespaceResponse, error) {
	request := heredoc.Doc(`
		query($name: String!){
          namespace(name: $name) {
            name
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
        }`)
	r := c.NewRequest(request)
	r.Var("name", name)

	var respData NamespaceResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return respData, err
	}

	return respData, nil
}
