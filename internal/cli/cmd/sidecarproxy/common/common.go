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
	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

func GetNamespaceNamesInteractively(cli cli.CLI) (string, error) {
	var namespaceName string

	client, err := common.GetGraphQLClient(cli)
	if err != nil {
		return namespaceName, errors.WrapIf(err, "could not get graphql client")
	}
	resp, err := client.GetNamespaces()
	if err != nil {
		return namespaceName, errors.WrapIf(err, "could not list namespaces")
	}
	defer client.Close()

	namespaceNameWithClusterSlice := make([]string, len(resp.Namespaces))
	for i, namespace := range resp.Namespaces {
		namespaceNameWithClusterSlice[i] = namespace.Name
	}

	err = survey.AskOne(&survey.Select{Message: "Namespace:", Options: namespaceNameWithClusterSlice}, &namespaceName, survey.WithValidator(survey.Required))
	if err != nil {
		return namespaceName, errors.WrapIf(err, "failed to select namespace")
	}
	return namespaceName, nil
}
