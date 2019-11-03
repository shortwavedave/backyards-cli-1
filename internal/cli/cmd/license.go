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

package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/licenses"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

const LicenseVersion = "1"

func NewLicenseCommand(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "license",
		Short:       "Shows Backyards license",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.HelperCommand},
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := licenses.Licenses.Open(fmt.Sprintf("LICENSE-v%s.txt", LicenseVersion))
			if err != nil {
				return err
			}

			b, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			fmt.Fprintf(cli.Out(), "%s", string(b))

			return nil
		},
	}

	return cmd
}
