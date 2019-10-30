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

	"emperror.dev/errors"
	"github.com/MakeNowJust/heredoc"
	corev1 "k8s.io/api/core/v1"
)

type DisableAutoSidecarInjectionRequest struct {
	Name string `json:"name"`
}

type DisableAutoSidecarInjectionResponse struct {
	NameSpaces []corev1.Namespace `json:"disableAutoSidecarInjection"`
}

func (c *client) DisableAutoSidecarInjection(req DisableAutoSidecarInjectionRequest) (DisableAutoSidecarInjectionResponse, error) {
	request := heredoc.Doc(`
		mutation disableAutoSidecarInjection($namespace: String!) {
			disableAutoSidecarInjection(namespace: $namespace){
				name
			}
		}`)

	r := c.NewRequest(request)
	r.Var("namespace", req.Name)

	// run it and capture the response
	var respData DisableAutoSidecarInjectionResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return respData, errors.WrapWithDetails(err, "could not disable auto sidecar injection", "namespace", req.Name)
	}

	return respData, nil
}
