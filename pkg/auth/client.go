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

package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"emperror.dev/errors"
	"k8s.io/client-go/rest"

	"github.com/banzaicloud/backyards-cli/pkg/servererror"
)

const defaultLoginTimeout = time.Second * 2

type Client interface {
	Login() (*Credentials, error)
}

type client struct {
	config *rest.Config
	url    string
}

type AuthenticationMode string

const (
	TokenAuth AuthenticationMode = "token"
	CertAuth  AuthenticationMode = "cert"
)

type RequestBody struct {
	Mode       AuthenticationMode `json:"mode"`
	Token      string             `json:"token,omitempty"`
	ClientCert struct {
		// base64 encoded client key
		Key string `json:"key"`
		// base64 encoded client cert
		Cert string `json:"cert"`
	} `json:"cert,omitempty"`
}

type Credentials struct {
	User struct {
		Name   string   `json:"name"`
		Groups []string `json:"groups"`
		// Token is an ID token containing user info and capabilities loaded at login
		Token string `json:"token"`
		// WrappedToken is a very short lifetime encrypted token that wraps the ID token.
		// It's for cases where the token must be exposed as HTTP GET parameters over a
		// secure connection where the token will available in access logs and/or browser
		// history which would mean a potential security risk.
		WrappedToken string `json:"wrappedToken"`
	} `json:"user"`
}

func NewClient(config *rest.Config, url string) Client {
	return &client{
		config: config,
		url:    url,
	}
}

func (c *client) Login() (*Credentials, error) {
	response, err := c.sendRequest(c.url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		parsedResponse := &servererror.Problem{}
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}
		err = json.NewDecoder(bytes.NewBuffer(responseBody)).Decode(parsedResponse)
		if err != nil {
			return nil, errors.WrapWithDetails(err, "invalid response", "response", string(responseBody))
		}
		if response.StatusCode < 500 {
			if parsedResponse.ErrorCode == servererror.AuthDisabledErrorCode {
				return nil, servererror.ErrAuthDisabled
			}
			return nil, errors.Errorf("invalid request %s: `%s`", response.Status, responseBody)
		}
		return nil, errors.Errorf("server error: %s `%s`", response.Status, responseBody)
	}

	parsedResponse := &Credentials{}
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return parsedResponse, errors.Wrap(err, "failed to read response body")
	}
	err = json.NewDecoder(bytes.NewBuffer(responseBody)).Decode(parsedResponse)
	if err != nil {
		return parsedResponse, errors.WrapWithDetails(err, "invalid response", "response", string(responseBody))
	}
	if parsedResponse.User.Name == "" {
		return nil, errors.New("invalid response")
	}
	return parsedResponse, nil
}

func (c *client) requestBody() (*RequestBody, error) {
	rb := &RequestBody{}
	// nolint ifElseChain
	if c.config.BearerToken != "" {
		rb.Mode = TokenAuth
		rb.Token = c.config.BearerToken
		return rb, nil
	} else if c.config.BearerTokenFile != "" {
		bearerToken, err := ioutil.ReadFile(c.config.BearerTokenFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load bearer token from %s", c.config.BearerTokenFile)
		}
		rb.Mode = TokenAuth
		rb.Token = string(bearerToken)
		return rb, nil
	} else if c.config.TLSClientConfig.CertFile != "" && c.config.TLSClientConfig.KeyFile != "" {
		cert, err := ioutil.ReadFile(c.config.TLSClientConfig.CertFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load client cert from %s", c.config.TLSClientConfig.CertFile)
		}
		key, err := ioutil.ReadFile(c.config.TLSClientConfig.KeyFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load client key from %s", c.config.TLSClientConfig.KeyFile)
		}
		rb.Mode = CertAuth
		rb.ClientCert.Cert = base64.StdEncoding.EncodeToString(cert)
		rb.ClientCert.Key = base64.StdEncoding.EncodeToString(key)
		return rb, nil
	} else if len(c.config.TLSClientConfig.CertData) > 0 && len(c.config.TLSClientConfig.KeyData) > 0 {
		rb.Mode = CertAuth
		rb.ClientCert.Cert = base64.StdEncoding.EncodeToString(c.config.TLSClientConfig.CertData)
		rb.ClientCert.Key = base64.StdEncoding.EncodeToString(c.config.TLSClientConfig.KeyData)
		return rb, nil
	} else if c.config.ExecProvider != nil || c.config.AuthProvider != nil {
		return rb, nil
	}
	return nil, errors.NewWithDetails("no credentials found in the provided config")
}

func (c *client) sendRequest(url string) (*http.Response, error) {
	transport, err := rest.TransportFor(c.config)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   defaultLoginTimeout,
	}
	rb, err := c.requestBody()
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	err = json.NewEncoder(b).Encode(rb)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode request")
	}
	return httpClient.Post(url, "application/json", b)
}
