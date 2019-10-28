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
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
)

var (
	commandUsageTemplate *template.Template
	templFuncs           = template.FuncMap{
		"formatAlias": func(s []string) string {
			if len(s) > 0 {
				return fmt.Sprintf("(aliases: %s)", strings.Join(s, ", "))
			}
			return ""
		},
		"rpad": func(s string, padding int) string {
			template := fmt.Sprintf("  %%-%ds", padding)
			return fmt.Sprintf(template, s)
		},
	}
)

func init() {
	commandUsage := `
Usage:
  {{.Cmd.UseLine}}

{{ if .Cmd.HasParent }}\
Commands:
{{range .Cmd.Commands }}\
{{ $formatedAliases := formatAlias .Aliases }}\
{{rpad .Name .NamePadding }} {{.Short}} {{ $formatedAliases }}
{{end}}\
{{end}}\
{{ if not .Cmd.HasParent }}\

Commands:
{{range .InstallCommands}}\
{{ $formatedAliases := formatAlias .Aliases }}\
{{ if .IsAvailableCommand }}\
{{rpad .Name .NamePadding }} {{.Short}} {{ $formatedAliases }}
{{end}}\
{{end}}\
{{end}}\
{{ if not .Cmd.HasParent }}\

{{range .OperationCommands}}\
{{ $formatedAliases := formatAlias .Aliases }}\
{{ if .IsAvailableCommand }}\
{{rpad .Name .NamePadding }} {{.Short}} {{ $formatedAliases }}
{{end}}\
{{end}}\
{{end}}\
{{ if not .Cmd.HasParent }}\

{{range .HelperCommands}}\
{{ $formatedAliases := formatAlias .Aliases }}\
{{ if .IsAvailableCommand }}\
{{rpad .Name .NamePadding }} {{.Short}} {{ $formatedAliases }}
{{end}}\
{{end}}\
{{end}}\
{{ if not .Cmd.HasParent }}\
{{range .Cmd.Commands }}\
{{if eq .Name "help"}}\
{{ $formatedAliases := formatAlias .Aliases }}\
{{rpad .Name .NamePadding }} {{.Short}} {{ $formatedAliases }}
{{end}}\
{{end}}\
{{end}}\
{{ if not .Cmd.HasParent }}\


Components commands:
{{range .ComponentCommands}}\
{{ $formatedAliases := formatAlias .Aliases }}\
{{ if .IsAvailableCommand }}\
{{rpad .Name .NamePadding }} {{.Short}} {{ $formatedAliases }}
{{end}}\
{{end}}\
{{end}}\
{{if .Cmd.HasLocalFlags}}\

Flags:
{{ .Cmd.LocalFlags.FlagUsages}}\
{{end}}\
{{if .Cmd.HasAvailableInheritedFlags}}\

GlobalFlags: 
{{ .Cmd.InheritedFlags.FlagUsages}}\
{{end}}\

Use "{{.Cmd.Name}} [command] --help" for more information about a command.
`[1:]

	commandUsageTemplate = template.Must(template.New("command_usage").Funcs(templFuncs).Parse(strings.Replace(commandUsage, "\\\n", "", -1)))
}

func getSubCommands(cmd *cobra.Command, cGroup string) []*cobra.Command {
	var subCommands []*cobra.Command
	for _, subCmd := range cmd.Commands() {
		subCmd := subCmd
		if g, ok := subCmd.Annotations[util.CommandGroupAnnotationKey]; ok && g == cGroup {
			subCommands = append(subCommands, subCmd)
		}
	}
	return subCommands
}

func usageFunc(cmd *cobra.Command) error {
	instCommands := getSubCommands(cmd, util.InstallCommand)
	opCommands := getSubCommands(cmd, util.OperationCommand)
	helpCommands := getSubCommands(cmd, util.HelperCommand)
	compCommands := getSubCommands(cmd, util.ComponentCommand)
	err := commandUsageTemplate.Execute(os.Stdout, struct {
		Cmd               *cobra.Command
		InstallCommands   []*cobra.Command
		OperationCommands []*cobra.Command
		HelperCommands    []*cobra.Command
		ComponentCommands []*cobra.Command
	}{
		Cmd:               cmd,
		InstallCommands:   instCommands,
		OperationCommands: opCommands,
		HelperCommands:    helpCommands,
		ComponentCommands: compCommands,
	})
	if err != nil {
		return err
	}
	return nil
}
