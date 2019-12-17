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

package mtls

import (
	"fmt"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/output"

	"github.com/banzaicloud/istio-client-go/pkg/authentication/v1alpha1"
)

type mTLSMode string

const (
	ModeStrict     mTLSMode = "STRICT"
	ModePermissive mTLSMode = "PERMISSIVE"
	ModeDisabled   mTLSMode = "DISABLED"
)

type Out struct {
	Policy   string   `json:"policy,omitempty"`
	Targets  []string `json:"targets,omitempty"`
	MtlsMode mTLSMode `json:"mtlsMode,omitempty"`
}

func Output(cli cli.CLI, resourceName types.NamespacedName, policySpec map[string][]*v1alpha1.PolicySpec) error {
	var err error

	outs := make([]Out, 0)
	for policyName, ps := range policySpec {
		for _, p := range ps {
			o := Out{}
			o.Targets = make([]string, len(p.Targets))
			for i, t := range p.Targets {
				o.Targets[i] = t.Name
				for _, p := range t.Ports {
					o.Targets[i] += "("
					if p.Name != nil {
						o.Targets[i] += *p.Name
					} else if p.Number != nil {
						o.Targets[i] += fmt.Sprint(*p.Number)
					}
					o.Targets[i] += ")"
				}
			}
			o.Policy = policyName
			switch {
			case p.Peers == nil:
				o.MtlsMode = ModeDisabled
			case p.Peers[0].Mtls == nil || p.Peers[0].Mtls.Mode == "" || p.Peers[0].Mtls.Mode == v1alpha1.ModeStrict:
				o.MtlsMode = ModeStrict
			case p.Peers[0].Mtls.Mode == v1alpha1.ModePermissive:
				o.MtlsMode = ModePermissive
			}

			outs = append(outs, o)
		}
	}

	if cli.OutputFormat() == output.OutputFormatTable && cli.Interactive() {
		fmt.Fprintf(cli.Out(), "mTLS rule for %s\n\n", resourceName)
	}

	err = show(cli, outs)
	if err != nil {
		return err
	}

	if cli.Interactive() {
		fmt.Println()
	}

	return nil
}

func show(cli output.FormatContext, data interface{}) error {
	ctx := &output.Context{
		Out:     cli.Out(),
		Color:   cli.Color(),
		Format:  cli.OutputFormat(),
		Fields:  []string{"Policy", "Targets", "MtlsMode"},
		Headers: []string{"Policy", "Targets", "MtlsMode"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		return errors.WrapIf(err, "could not produce output")
	}

	return nil
}
