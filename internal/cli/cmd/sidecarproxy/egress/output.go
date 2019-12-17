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

func Output(cli cli.CLI, workloadName types.NamespacedName, sidecars []graphql.Sidecar, recommendation bool) error {
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
