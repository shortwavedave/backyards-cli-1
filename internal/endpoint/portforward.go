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

package endpoint

import (
	"net/http"

	"github.com/banzaicloud/backyards-cli/pkg/k8s/portforward"
)

type portForwardEndpoint struct {
	pf *portforward.Portforward
	ca []byte
}

func NewPortforwardEndpoint(pf *portforward.Portforward, ca []byte) Endpoint {
	return &portForwardEndpoint{
		pf: pf,
		ca: ca,
	}
}

func (e *portForwardEndpoint) URLForPath(path string) string {
	return e.pf.GetURL(path)
}

func (e *portForwardEndpoint) CA() []byte {
	return e.ca
}

func (e *portForwardEndpoint) HTTPClient() *http.Client {
	return withCa(e.ca)
}

func (e *portForwardEndpoint) Close() {
	e.pf.Stop()
}
