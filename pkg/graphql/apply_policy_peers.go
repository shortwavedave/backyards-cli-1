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

	"github.com/MakeNowJust/heredoc"
)

type ApplyPolicyPeersInput struct {
	Selector *PolicySelectorInput             `json:"selector"`
	Peers    []*PeerAuthenticationMethodInput `json:"peers"`
}

type PolicySelectorInput struct {
	Namespace string               `json:"namespace"`
	Target    *TargetSelectorInput `json:"target"`
}

type TargetSelectorInput struct {
	Name string                 `json:"name"`
	Port *AuthPortSelectorInput `json:"port"`
}

type AuthPortSelectorInput struct {
	Number *int    `json:"number"`
	Name   *string `json:"name"`
}

type PeerAuthenticationMethodInput struct {
	Mtls *MutualTLSInput `json:"mtls"`
}

type MutualTLSInput struct {
	Mode *AuthTLSModeInput `json:"mode"`
}

type AuthTLSModeInput string

const (
	AuthTLSModeInputStrict     AuthTLSModeInput = "STRICT"
	AuthTLSModeInputPermissive AuthTLSModeInput = "PERMISSIVE"
)

func AuthTLSModeInputToPointer(mode AuthTLSModeInput) *AuthTLSModeInput {
	return &mode
}

func (c *client) ApplyPolicyPeers(input ApplyPolicyPeersInput) (bool, error) {
	request := heredoc.Doc(`
	  mutation applyPolicyPeers(
        $input: ApplyPolicyPeersInput!
      ) {
        applyPolicyPeers(
          input: $input
        )
      }
`)

	r := c.NewRequest(request)
	r.Var("input", input)

	// run it and capture the response
	var respData map[string]bool
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return false, err
	}

	return respData["applyPolicyPeers"], nil
}
