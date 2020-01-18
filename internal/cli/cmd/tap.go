// Copyright © 2020 Banzai Cloud
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
	"context"
	"net/http"
	"strings"
	"text/template"

	"emperror.dev/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/ale"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/output"
)

const (
	resourceTypeWorkload = "WORKLOAD"
	resourceTypePod      = "POD"
)

type tapCommand struct {
	cli cli.CLI
}

type res struct {
	Type      string
	Name      string
	Namespace string
}

type TapOptions struct {
	namespace            string
	reporter             res
	destination          res
	destinationResource  string
	destinationNamespace string
	authority            string
	method               string
	path                 string
	scheme               string
	direction            string
	responseCode         []uint
	responseCodeMin      uint
	responseCodeMax      uint
}

func NewTapOptions() *TapOptions {
	return &TapOptions{
		namespace: "default",
	}
}

func NewTapCmd(cli cli.CLI, options *TapOptions) *cobra.Command {
	c := &tapCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "tap [[ns|workload|pod]/resource-name]",
		Short: "Tap into HTTP/GRPC mesh traffic",
		Example: `
  # tap the movies-v1 deployment in the default namespace
  backyards tap workload/movies-v1

  # tap the movies-v1-7f9645bfd7-8vgm9 pod in the default namespace
  backyards tap pod/movies-v1-7f9645bfd7-8vgm9

  # tap the backyards-demo namespace
  backyards tap ns/backyards-demo

  # tap the backyards-demo namespace request to test namespace
  backyards tap ns/backyards-demo --destination ns/test`,
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.OperationCommand},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := c.parseResource(args[0], &options.reporter)
			if err != nil {
				return errors.WrapIf(err, "invalid resource")
			}
			if options.reporter.Namespace != "" {
				options.namespace = options.reporter.Namespace
			}

			if options.destinationResource != "" {
				err = c.parseResource(options.destinationResource, &options.destination)
				if err != nil {
					return errors.WrapIf(err, "invalid destination resource")
				}
				if options.destination.Namespace != "" {
					options.destinationNamespace = options.destination.Namespace
				}
			}

			if len(options.responseCode) == 1 {
				options.responseCodeMin = options.responseCode[0]
				options.responseCodeMax = options.responseCode[0]
			} else if len(options.responseCode) >= 2 {
				options.responseCodeMin = options.responseCode[0]
				options.responseCodeMax = options.responseCode[1]
			}

			if options.responseCodeMin > options.responseCodeMax {
				options.responseCodeMax, options.responseCodeMin = options.responseCodeMin, options.responseCodeMax
			}

			options.method = strings.ToUpper(options.method)
			options.direction = strings.ToUpper(options.direction)

			return c.run(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.namespace, "ns", options.namespace, "Namespace of the specified resource")
	cmd.Flags().StringVar(&options.destinationNamespace, "destination-ns", options.destinationNamespace, "Namespace of the destination resource; by default the current \"--namespace\" is used")
	cmd.Flags().StringVar(&options.destinationResource, "destination", options.destinationResource, "Show requests to this resource")

	cmd.Flags().StringVar(&options.path, "path", options.path, "Show requests with paths with this prefix")
	cmd.Flags().StringVar(&options.scheme, "scheme", options.scheme, "Show requests with this scheme")
	cmd.Flags().StringVar(&options.method, "method", options.method, "Show requests with this request method")
	cmd.Flags().StringVar(&options.authority, "authority", options.authority, "Show requests with this authority")
	cmd.Flags().StringVar(&options.direction, "direction", options.direction, "Show requests with this direction (inbound|outbound)")

	cmd.Flags().UintSliceVar(&options.responseCode, "response-code", options.responseCode, "Show request with this response code")

	return cmd
}

func (c *tapCommand) run(cli cli.CLI, options *TapOptions) error {
	var err error

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	err = c.validateOptions(client, options)
	if err != nil {
		return err
	}

	input := &graphql.GetAccessLogsInput{
		ReporterNamespace:    options.reporter.Namespace,
		DestinationNamespace: options.destination.Namespace,
		Authority:            options.authority,
		Method:               options.method,
		Path:                 options.path,
		Scheme:               options.scheme,
		Direction:            options.direction,
		StatusCode: graphql.IntRange{
			Min: options.responseCodeMin,
			Max: options.responseCodeMax,
		},
	}
	if options.reporter.Type != "" {
		input.ReporterType = options.reporter.Type
		input.ReporterName = options.reporter.Name
	}
	input.ReporterNamespace = options.namespace
	if options.destination.Type != "" {
		input.DestinationType = options.destination.Type
		input.DestinationName = options.destination.Name
	}

	ch := make(chan interface{})
	errCh := make(chan error)
	ctx, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()
	go func() {
		client.SubscribeToAccessLogs(ctx, input, ch, errCh)
	}()

	type Data map[string]interface{}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			return err
		case msg := <-ch:
			var data Data
			err := mapstructure.Decode(msg, &data)
			if err != nil {
				return errors.WrapIf(err, "could not parse message")
			}

			var entry ale.HTTPAccessLogEntry
			err = mapstructure.Decode(data["accessLogs"], &entry)
			if err != nil {
				return errors.WrapIf(err, "could not parse message")
			}

			tpl, _ := template.New("format").Parse(`{{.StartTime}} {{.Direction}} {{.Source}} {{.Destination}} "{{.Request.Scheme}} {{.Request.Method}} {{.Request.Path}} {{.ProtocolVersion}}" {{.Response.StatusCode}} {{.Latency}} "{{.Destination.Address}}"`)

			switch cli.OutputFormat() {
			case output.OutputFormatJSON, output.OutputFormatYAML:
				err = output.Output(&output.Context{
					Out:    cli.Out(),
					Color:  cli.Color(),
					Format: cli.OutputFormat()}, entry)
			default:
				_, err = cli.Out().Write([]byte(entry.FormattedString(tpl) + "\n"))
			}

			if err != nil {
				return err
			}
		}
	}
}

func (c *tapCommand) parseResource(res string, parsed *res) error {
	parts := strings.Split(res, "/")
	if len(parts) < 2 {
		return errors.Errorf("invalid resource, supported format: [ns|workload|pod]/resource-name")
	}
	if !util.IsValidK8sResourceName(parts[1]) {
		return errors.Errorf("invalid resource name: %s", parts[1])
	}

	parsed.Name = parts[1]

	switch parts[0] {
	case "workload":
		parsed.Type = resourceTypeWorkload
	case "pod":
		parsed.Type = resourceTypePod
	case "ns":
		parsed.Namespace = parts[1]
	default:
		return errors.Errorf("invalid resource type: %s, use workload, pod or ns", parts[0])
	}

	return nil
}

func (c *tapCommand) validateOptions(client graphql.Client, options *TapOptions) error {
	ns, err := client.GetNamespace(options.namespace)
	if err != nil {
		return err
	}

	if ns.Namespace.Name == "" {
		return errors.Errorf("invalid namespace: %s", options.namespace)
	}

	switch options.reporter.Type {
	case resourceTypePod:
		_, err = client.GetPod(options.namespace, options.reporter.Name)
		if err != nil {
			return errors.WrapIff(err, "invalid pod: %s/%s", options.namespace, options.reporter.Name)
		}
	case resourceTypeWorkload:
		_, err = client.GetWorkload(options.namespace, options.reporter.Name)
		if err != nil {
			return errors.WrapIff(err, "invalid workload: %s/%s", options.namespace, options.reporter.Name)
		}
	}

	if options.destinationNamespace != "" && options.destinationNamespace != options.namespace {
		ns, err := client.GetNamespace(options.destinationNamespace)
		if err != nil {
			return err
		}

		if ns.Namespace.Name == "" {
			return errors.Errorf("invalid namespace: %s", options.destinationNamespace)
		}
	}

	switch options.destination.Type {
	case resourceTypePod:
		_, err = client.GetPod(options.destinationNamespace, options.destination.Name)
		if err != nil {
			return errors.WrapIff(err, "invalid pod: %s/%s", options.destinationNamespace, options.destination.Name)
		}
	case resourceTypeWorkload:
		_, err = client.GetWorkload(options.destinationNamespace, options.destination.Name)
		if err != nil {
			return errors.WrapIff(err, "invalid workload: %s/%s", options.destinationNamespace, options.destination.Name)
		}
	}

	switch options.direction {
	case "IN":
		options.direction = "INBOUND"
	case "OUT":
		options.direction = "OUTBOUND"
	case "", "INBOUND", "OUTBOUND":
	default:
		return errors.Errorf("invalid traffic direction: %s", options.direction)
	}

	switch options.method {
	case "",
		http.MethodConnect, http.MethodDelete, http.MethodGet,
		http.MethodHead, http.MethodOptions, http.MethodPatch,
		http.MethodPost, http.MethodPut, http.MethodTrace:
	default:
		return errors.Errorf("invalid HTTP method: %s", options.method)
	}

	return nil
}
