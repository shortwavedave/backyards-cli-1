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

type MeshService struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`

	VirtualServices  []v1alpha3.VirtualService  `json:"virtualServices"`
	DestinationRules []v1alpha3.DestinationRule `json:"destinationRules"`
	Policies         []Policy                   `json:"policies"`
}

func (c *client) GetService(namespace, name string) (*MeshService, error) {
	request := heredoc.Doc(`
	query service($name:String!, $namespace: String!) {
		service(name:$name, namespace: $namespace){
		  id
		  name
		  namespace
		  destinationRules {
			spec {
				trafficPolicy {
				  outlierDetection {
					interval
					baseEjectionTime
					consecutiveErrors
					maxEjectionPercent
				  }
				connectionPool {
				  tcp {
				    maxConnections
					connectTimeout
				  }
				  http {
				    maxRetries
				    http1MaxPendingRequests
					http2MaxRequests
					maxRequestsPerConnection
				  }
				}
			  }
			}
		  }
		  virtualServices {
			spec {
			  exportTo
			  hosts
			  http {
				timeout
				fault {
					delay {
					  percentage {
						value
					  }
					  fixedDelay
					}
					abort {
					  percentage {
						value
					  }
					  httpStatus
					}
				  }
				route {
				  weight
				  destination {
					host
					subset
					port {
					  number
					}
				  }
				}
				redirect {
					uri
					authority
					redirectCode
				}
				retries {
					attempts
					perTryTimeout
					retryOn
				}
				rewrite {
					uri
					authority
				}
				mirror {
					host
					subset
					port {
						number
					}
				}
				match {
				  name
				  uri {
					suffix
					prefix
					regex
					exact
				  }
				  method {
					suffix
					prefix
					regex
					exact
				  }
				  authority {
					suffix
					prefix
					regex
					exact
				  }
				  scheme {
					suffix
					prefix
					regex
					exact
				  }
				  sourceLabels
				}
			  }
			}
			metadata {
			  uid
			  name
			  namespace
			}
		  }
		}
	  }
`)

	type Response struct {
		Service MeshService `json:"service"`
	}

	r := c.NewRequest(request)
	r.Var("name", name)
	r.Var("namespace", namespace)

	// run it and capture the response
	var respData Response
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return nil, err
	}

	if respData.Service.ID == "" {
		return nil, errors.New("not found")
	}

	return &respData.Service, nil
}

func (c *client) GetServiceWithMTLS(namespace, name string) (*MeshService, error) {
	request := heredoc.Doc(`
	query service($name:String!, $namespace: String!) {
		service(name:$name, namespace: $namespace){
		  id
		  name
		  namespace
		  policies {
		    name
		    namespace
		    spec {
			  targets {
			    name
			    ports {
				  name
				  number
			    }
			  }
			  peers {
			    mtls {
				  mode
			    }
			  }
		    }
		  }
        }
      }
`)

	type Response struct {
		Service MeshService `json:"service"`
	}

	r := c.NewRequest(request)
	r.Var("name", name)
	r.Var("namespace", namespace)

	// run it and capture the response
	var respData Response
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return nil, err
	}

	if respData.Service.ID == "" {
		return nil, errors.New("not found")
	}

	return &respData.Service, nil
}
