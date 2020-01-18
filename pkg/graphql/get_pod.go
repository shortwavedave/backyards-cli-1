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
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/MakeNowJust/heredoc"
)

type Pod struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func (c *client) GetPod(namespace, name string) (*Pod, error) {
	request := heredoc.Doc(`
	query namespace($name: String!) {
	  node(id: $name) {
		__typename
		... on Pod {
		  name
		  namespace
		}
	  }
	}`)
	r := c.NewRequest(request)

	clusters, err := c.Clusters()
	if err != nil {
		return nil, err
	}

	type Response struct {
		Pod Pod `json:"node"`
	}

	var respData Response
	for _, cluster := range clusters {
		clusterName := cluster.Name
		if cluster.Type == "host" {
			clusterName = "master"
		}
		name := fmt.Sprintf("pod:%s:%s:%s", clusterName, namespace, name)
		r.Var("name", name)
		if err := c.client.Run(context.Background(), r, &respData); err != nil && !strings.Contains(err.Error(), "not found") {
			return nil, err
		}
	}

	if respData.Pod.Name != name && respData.Pod.Namespace != namespace {
		return nil, errors.Errorf("not found")
	}

	return &respData.Pod, nil
}
