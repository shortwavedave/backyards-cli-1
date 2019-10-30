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

package fi

import (
	"fmt"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/output"
	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

type Out struct {
	Matches         string  `json:"matches,omitempty" yaml:"matchers,omitempty"`
	DelayPercentage float32 `json:"delayPercentage,omitempty" yaml:"delayPercentage,omitempty"`
	FixedDelay      string  `json:"fixedDelay,omitempty" yaml:"fixedDelay,omitempty"`
	AbortPercentage float32 `json:"abortPercentage,omitempty" yaml:"abortPercentage,omitempty"`
	AbortStatusCode int     `json:"abortStatusCode,omitempty" yaml:"abortStatusCode,omitempty"`
}

func Output(cli cli.CLI, serviceName types.NamespacedName, routes ...v1alpha3.HTTPRoute) error {
	var o Out
	var err error

	outs := make([]Out, 0)
	for _, route := range routes {
		if route.Fault != nil {
			if route.Fault.Delay != nil {
				o.DelayPercentage = route.Fault.Delay.Percentage.Value
				o.FixedDelay = route.Fault.Delay.FixedDelay
			}
			if route.Fault.Abort != nil {
				o.AbortPercentage = route.Fault.Abort.Percentage.Value
				o.AbortStatusCode = route.Fault.Abort.HTTPStatus
			}
		}
		o.Matches = common.HTTPMatchRequests(common.ConvertHTTPMatchRequestsPointers(route.Match)).String()
		outs = append(outs, o)
	}

	if cli.OutputFormat() == output.OutputFormatTable && cli.Interactive() {
		fmt.Printf("Fault injection settings for %s\n\n", serviceName)
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
		Fields:  []string{"Matches", "DelayPercentage", "FixedDelay", "AbortPercentage", "AbortStatusCode"},
		Headers: []string{"Matches", "Delay percentage", "Fixed delay", "Abort percentage", "Abort http status code"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		return errors.WrapIf(err, "could not produce output")
	}

	return nil
}
