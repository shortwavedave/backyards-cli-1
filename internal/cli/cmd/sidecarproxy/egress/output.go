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

package egress

import (
	"bytes"
	"fmt"

	"github.com/banzaicloud/backyards-cli/pkg/graphql"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy/common"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/output"
)

type Out struct {
	Sidecar     string      `json:"sidecar,omitempty"`
	Selector    string      `json:"selector:omitempty"`
	Hosts       string      `json:"hosts,omitempty"`
	Port        common.Port `json:"port,omitempty"`
	Bind        string      `json:"bind,omitempty"`
	CaptureMode string      `json:"capture_mode,omitempty"`
}

func Output(cli cli.CLI, workloadName types.NamespacedName, sidecars []graphql.Sidecar, recommendation, apply bool) error {
	var err error

	outs := make([]Out, 0)
	for _, sc := range sidecars {
		var selector string
		if sc.Spec.WorkloadSelector != nil {
			b := new(bytes.Buffer)
			for key, value := range sc.Spec.WorkloadSelector.Labels {
				fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
			}
			selector = b.String()
		}
		for _, e := range sc.Spec.Egress {
			var hosts string
			b := new(bytes.Buffer)
			for _, h := range e.Hosts {
				fmt.Fprintf(b, "%s\n", h)
			}
			hosts = b.String()

			o := Out{}
			o.Sidecar = sc.Name
			o.Selector = selector
			o.Hosts = hosts
			o.Bind = e.Bind
			if e.Port != nil {
				o.Port = common.Port(*e.Port)
			}
			o.CaptureMode = string(e.CaptureMode)

			outs = append(outs, o)
		}
	}

	if len(outs) == 0 {
		if recommendation {
			fmt.Fprintf(cli.Out(), "no recommended egress rule found for %s\n\n", workloadName)
		} else {
			fmt.Fprintf(cli.Out(), "no egress rule found for %s\n\n", workloadName)
		}
		return nil
	}

	if cli.OutputFormat() == output.OutputFormatTable && cli.Interactive() {
		if recommendation {
			fmt.Fprintf(cli.Out(), "Recommended sidecar egress rules for %s\n\n", workloadName)
		} else {
			fmt.Fprintf(cli.Out(), "Sidecar egress rules for %s\n\n", workloadName)
		}
	}

	err = show(cli, outs)
	if err != nil {
		return err
	}

	if recommendation {
		// recommendation shouldn't contain more than 1 sidecar, and we don't want to print a hint when there's no recommendation
		if len(sidecars) == 1 {
			printRecommendationHint(cli, workloadName, sidecars[0], apply)
		}
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
		Fields:  []string{"Sidecar", "Selector", "Hosts", "Bind", "Port", "CaptureMode"},
		Headers: []string{"Sidecar", "Selector", "Hosts", "Bind", "Port", "Capture Mode"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		return errors.WrapIf(err, "could not produce output")
	}

	return nil
}

func printRecommendationHint(cli output.FormatContext, workloadName types.NamespacedName, sidecar graphql.Sidecar, apply bool) {
	var hosts []string
	for _, e := range sidecar.Spec.Egress {
		// recommendations are always for egress without bind and port
		if e.Port == nil && e.Bind == "" {
			hosts = e.Hosts
			break
		}
	}

	// couldn't find recommended hosts, we don't print anything
	if len(hosts) == 0 {
		return
	}

	var applyCommand = fmt.Sprintf("> backyards sp egress set --workload %s/%s", workloadName.Namespace, workloadName.Name)
	for _, h := range hosts {
		applyCommand += fmt.Sprintf(" --hosts=%s", h)
	}
	if sidecar.Spec.WorkloadSelector != nil {
		for l := range sidecar.Spec.WorkloadSelector.Labels {
			applyCommand += fmt.Sprintf(" -l=%s", l)
		}
	}
	var hint string
	if apply {
		hint = fmt.Sprintf("\nHint: use this command to apply these recommendations manually:\n"+
			"%s\n\n", applyCommand)
	} else {
		hint = fmt.Sprintf("\nHint: to apply these recommendations, use the --apply switch, or apply it manually using this command:\n"+
			"%s\n\n", applyCommand)

	}
	fmt.Fprintf(cli.Out(), "%s", hint)
}
