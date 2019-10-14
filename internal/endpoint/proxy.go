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
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"emperror.dev/errors"
	"k8s.io/client-go/rest"
)

type K8sService struct {
	Name      string
	Namespace string
	Port      int
}

func (s K8sService) Path() string {
	return fmt.Sprintf("/api/v1/namespaces/%s/services/%s:%d/proxy", s.Namespace, s.Name, s.Port)
}

type proxyEndpoint struct {
	baseURL string
	service K8sService
	ca      []byte

	srv *http.Server
}

func NewProxyEndpoint(localPort int, cfg *rest.Config, service K8sService) (Endpoint, error) {
	var err error

	if localPort == 0 {
		localPort, err = getEphemeralPort()
		if err != nil {
			return nil, errors.WrapIf(err, "could not get ephemeral port")
		}
	}

	hostPort := fmt.Sprintf("127.0.0.1:%d", localPort)

	ep := &proxyEndpoint{
		service: service,
		baseURL: "http://" + hostPort,
	}

	mux := http.NewServeMux()
	mux.Handle("/", ep.proxyToCluster(cfg))

	ep.srv = &http.Server{
		Addr:    hostPort,
		Handler: mux,
	}

	go func() {
		_ = ep.srv.ListenAndServe()
	}()

	return ep, nil
}

func (e *proxyEndpoint) URLForPath(path string) string {
	return fmt.Sprintf("%s%s", e.baseURL, path)
}

func (e *proxyEndpoint) CA() []byte {
	return e.ca
}

func (e *proxyEndpoint) HTTPClient() *http.Client {
	return withCa(e.ca)
}

func (e *proxyEndpoint) Close() {
	_ = e.srv.Shutdown(context.Background())
}

func (e *proxyEndpoint) proxyToCluster(cfg *rest.Config) http.Handler {
	proxyPath := e.service.Path()
	h, _ := NewK8sAPIProxy(cfg, proxyPath)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = proxyPath + r.URL.Path
		authHeaderValue := r.Header.Get("Authorization")
		if authHeaderValue != "" {
			r.Header.Set("X-Authorization", authHeaderValue)
		}
		h.ServeHTTP(w, r)
	})
}

func NewK8sAPIProxy(cfg *rest.Config, proxyPath string) (http.Handler, error) {
	host := cfg.Host
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	target, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	transport, err := rest.TransportFor(cfg)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: target.Scheme, Host: target.Host})
	proxy.Transport = &ReplaceTransport{
		PathPrepend:  proxyPath,
		RoundTripper: transport,
	}
	proxy.FlushInterval = 200 * time.Millisecond

	return http.Handler(proxy), nil
}

// getEphemeralPort selects a port for listening on
// It binds to a free ephemeral port and returns the port number
func getEphemeralPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, errors.WrapIf(err, "could not listen on port zero")
	}

	defer listener.Close()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.NewWithDetails("invalid listen address", "address", listener.Addr())
	}

	return tcpAddr.Port, nil
}
