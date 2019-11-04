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
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/login"
	"github.com/banzaicloud/backyards-cli/pkg/auth"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

func GetGraphQLClient(cli cli.CLI) (graphql.Client, error) {
	token := cli.GetToken()
	if token == "" {
		err := login.Login(cli, func(body *auth.Credentials) {
			token = body.User.Token
		})
		if err != nil {
			return nil, err
		}
	}

	endpoint, err := cli.InitializedEndpoint()
	if err != nil {
		return nil, err
	}

	client := graphql.NewClient(endpoint, "/api/graphql")

	if token != "" {
		client.SetJWTToken(token)
	}

	return client, nil
}
