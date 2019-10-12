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
	"fmt"
	"net/http"
)

type externalEndpoint struct {
	baseURL string
	ca      []byte
}

func NewExternalEndpoint(baseURL string, ca []byte) Endpoint {
	return &externalEndpoint{
		baseURL: baseURL,
		ca:      ca,
	}
}

func (e *externalEndpoint) URLForPath(path string) string {
	return fmt.Sprintf("%s%s", e.baseURL, path)
}

func (e *externalEndpoint) CA() []byte {
	return e.ca
}

func (e *externalEndpoint) HTTPClient() *http.Client {
	return withCa(e.ca)
}

func (e *externalEndpoint) Close() {
}
