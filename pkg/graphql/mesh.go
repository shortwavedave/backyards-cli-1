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

	"github.com/banzaicloud/istio-client-go/pkg/authentication/v1alpha1"
)

type MeshPolicy struct {
	v1alpha1.MeshPolicy
	Name string `json:"name"`
}

type MeshWithPolicy struct {
	MeshPolicy *MeshPolicy `json:"meshPolicy"`
}

func (c *client) GetMeshWithMTLS() (*MeshPolicy, error) {
	request := heredoc.Doc(`
		query mesh {
		  mesh(namespaces: []) {
			meshPolicy {
			  name
			  spec {
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
		Mesh MeshWithPolicy `json:"mesh"`
	}

	r := c.NewRequest(request)

	// run it and capture the response
	var respData Response
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return nil, err
	}

	if respData.Mesh.MeshPolicy == nil {
		return nil, errors.New("not found")
	}

	return respData.Mesh.MeshPolicy, nil
}
