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
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy/common"

	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

type recommendCommand struct{}

type RecommendOptions struct {
	workloadID     string
	workloadName   types.NamespacedName
	isolationLevel string
}

func NewRecommendOptions() *RecommendOptions {
	return &RecommendOptions{}
}

func NewRecommendCommand(cli cli.CLI) *cobra.Command {
	c := &recommendCommand{}
	options := NewRecommendOptions()

	cmd := &cobra.Command{
		Use:           "recommend [[--workload=]namespace/workloadname] [--isolationLevel=NAMESPACE|WORKLOAD]",
		Short:         "Recommend sidecar configuration for a workload",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) > 0 {
				options.workloadID = args[0]
			}

			if options.workloadID == "" {
				return errors.New("workload must be specified")
			}

			options.workloadName, err = util.ParseK8sResourceIDAllowWildcard(options.workloadID)
			if err != nil {
				return errors.WrapIf(err, "could not parse workload ID")
			}

			err = c.validateOptions(options)
			if err != nil {
				return errors.WithStack(err)
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.workloadID, "workload", "", "Workload name")
	flags.StringVarP(&options.isolationLevel, "isolationLevel", "i", "", "Isolation level (NAMESPACE|WORKLOAD)")

	return cmd
}

func (c *recommendCommand) validateOptions(options *RecommendOptions) error {

	if len(options.isolationLevel) != 0 && options.isolationLevel != "WORKLOAD" && options.isolationLevel != "NAMESPACE" {
		return errors.New("isolation level must be one of WORKLOAD or NAMESPACE")
	}

	return nil
}

func (c *recommendCommand) run(cli cli.CLI, options *RecommendOptions) error {
	var err error

	client, err := cmdCommon.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	var egressListeners map[string][]*v1alpha3.IstioEgressListener
	if options.workloadName.Name != "*" {
		wl, err := client.GetWorkloadWithSidecarRecommendation(options.workloadName.Namespace, options.workloadName.Name, options.isolationLevel, nil)
		if err != nil {
			return errors.Wrap(err, "couldn't query workload sidecar recommendations")
		}

		if len(wl.RecommendedSidecars) == 0 {
			log.Infof("no sidecar recommendations found for %s", options.workloadName)
			return nil
		}

		egressListeners = common.GetEgressListenerMap(wl.RecommendedSidecars)
	} else {
		resp, err := client.GetNamespaceWithSidecarRecommendation(options.workloadName.Namespace, options.isolationLevel)
		if err != nil {
			return errors.Wrap(err, "couldn't query namespace sidecar recommendations")
		}
		egressListeners = common.GetEgressListenerMap(resp.Namespace.RecommendedSidecars)
	}

	if len(egressListeners) == 0 {
		log.Infof("no recommended egress rule found for %s", options.workloadName)
		return nil
	}

	return Output(cli, options.workloadName, egressListeners, true)
}
