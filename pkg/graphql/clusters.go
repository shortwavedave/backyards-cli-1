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

type ClusterStatus struct {
	ErrorMessage   string   `json:"errorMessage"`
	GatewayAddress []string `json:"gatewayAddress"`
	Status         string   `json:"status"`
}

type Cluster struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Type      string        `json:"type"`
	Status    ClusterStatus `json:"status"`
}

type ClustersResponse []Cluster

func (cr ClustersResponse) GetHostCluster() (bool, *Cluster) {
	for _, c := range cr {
		c := c
		if c.Type == "host" {
			return true, &c
		}
	}

	return false, nil
}

func (cr ClustersResponse) GetClusterByName(name string) (bool, *Cluster) {
	for _, c := range cr {
		c := c
		if c.Name == name {
			return true, &c
		}
	}

	return false, nil
}

func (c *client) Clusters() (ClustersResponse, error) {
	request := heredoc.Doc(`
	{
	  clusters {
		id
		name
		namespace
		type
		... on HostCluster {
			status {
			errorMessage
			gatewayAddress
			status
		  }
		}
		... on PeerCluster {
			status {
			errorMessage
			gatewayAddress
			status
		  }
		}
	  }
	}`)

	r := c.NewRequest(request)

	// run it and capture the response
	var respData map[string]ClustersResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return ClustersResponse{}, err
	}

	return respData["clusters"], nil
}
