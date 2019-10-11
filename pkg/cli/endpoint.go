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

package cli

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/banzaicloud/backyards-cli/pkg/k8s/portforward"
)

type Endpoint interface {
	URLForPath(path string) string
	CA() []byte
	HTTPClient() *http.Client
	Close()
}

type externalEndpoint struct {
	baseURL string
	ca      []byte
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

type portForwardEndpoint struct {
	pf *portforward.Portforward
	ca []byte
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

func withCa(ca []byte) *http.Client {
	if len(ca) > 0 {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		rootCAs.AppendCertsFromPEM(ca)
		tlsConfig := &tls.Config{
			RootCAs: rootCAs,
		}
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
	}
	return http.DefaultClient
}

func (e *portForwardEndpoint) Close() {
	e.pf.Stop()
}
