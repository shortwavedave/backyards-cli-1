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

package common

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"regexp"
	"strings"

	"emperror.dev/errors"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis/istio/v1alpha3"

	"github.com/banzaicloud/backyards-cli/pkg/auth"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

const (
	backyardsServiceAccountName        = "backyards"
	dns1123LabelFmt             string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
)

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

func ParseServiceID(serviceID string) (types.NamespacedName, error) {
	parts := strings.Split(serviceID, "/")
	if len(parts) != 2 {
		return types.NamespacedName{}, errors.Errorf("invalid service ID: '%s': format must be <namespace>/<name>", serviceID)
	}

	for _, p := range parts {
		if !dns1123LabelRegexp.MatchString(p) {
			return types.NamespacedName{}, errors.Errorf("invalid service ID: '%s': format must be <namespace>/<name>", serviceID)
		}
	}

	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}

func GetServiceByName(cli cli.CLI, serviceName types.NamespacedName) (*corev1.Service, error) {
	var service corev1.Service

	k8sclient, err := cli.GetK8sClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = k8sclient.Get(context.Background(), serviceName, &service)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &service, nil
}

func GetDestinationRuleByName(cli cli.CLI, serviceName types.NamespacedName) (*v1alpha3.DestinationRule, error) {
	var drule v1alpha3.DestinationRule

	k8sclient, err := cli.GetK8sClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = k8sclient.Get(context.Background(), serviceName, &drule)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &drule, nil
}

func GetVirtualserviceByName(cli cli.CLI, serviceName types.NamespacedName) (*v1alpha3.VirtualService, error) {
	var vservice v1alpha3.VirtualService

	k8sclient, err := cli.GetK8sClient()
	if err != nil {
		return nil, err
	}

	err = k8sclient.Get(context.Background(), serviceName, &vservice)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get virtual service")
	}

	return &vservice, nil
}

func GetGraphQLClient(cli cli.CLI) (graphql.Client, error) {
	k8sclient, err := cli.GetK8sClient()
	if err != nil {
		return nil, err
	}

	token, err := k8s.GetTokenForServiceAccountName(k8sclient, types.NamespacedName{
		Name:      backyardsServiceAccountName,
		Namespace: viper.GetString("backyards.namespace"),
	})
	if err != nil {
		return nil, err
	}

	url, err := cli.GetEndpointURL("/api/graphql")
	if err != nil {
		return nil, err
	}

	client := graphql.NewClient(url)
	client.SetJWTToken(token)

	return client, nil
}

func GetAuthClient(cli cli.CLI) (auth.Client, error) {
	config, err := cli.GetK8sConfig()
	if err != nil {
		return nil, err
	}

	url, err := cli.GetEndpointURL("/api/login")
	if err != nil {
		return nil, err
	}

	var client *http.Client
	ca, err := cli.GetEndpointCA()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get endpoint CA")
	}
	if ca != nil {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		rootCAs.AppendCertsFromPEM(ca)
		tlsConfig := &tls.Config{
			RootCAs: rootCAs,
		}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: transport}
	} else {
		client = http.DefaultClient
	}

	return auth.NewClient(config, client, url), nil
}
