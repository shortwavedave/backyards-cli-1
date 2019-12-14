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

package zookeeper

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

var (
	backoff = wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    50,
	}
	zookeeperNamespace = "zookeeper"
)

type installCommand struct {
	cli cli.CLI
}

type InstallOptions struct {
	DumpResources bool

	namespace string
}

func NewInstallOptions(namespace string) *InstallOptions {
	return &InstallOptions{
		namespace: zookeeperNamespace,
	}
}

func NewInstallCommand(cli cli.CLI, options *InstallOptions) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Install zookeeper cluster",
		Long: `Installs zookeeper cluster.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(cli, options)
		},
	}

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *InstallOptions) error {
	objects, err := GetK8sObjects(options.namespace)
	if err != nil {
		return err
	}

	crds := make(object.K8sObjects, 0)
	objs := make(object.K8sObjects, 0)
	for _, obj := range objects {
		if obj.Kind == "CustomResourceDefinition" {
			crds = append(crds, obj)
		} else {
			objs = append(objs, obj)
		}
	}

	if !options.DumpResources {
		client, err := c.cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.ApplyCRDs(client, cli.LabelManager(), crds)
		if err != nil {
			return err
		}

		client, err = c.cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.ApplyResourceObjects(client, cli.LabelManager(), objs)
		if err != nil {
			return err
		}

		err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(objs, "ZookeeperCluster"), backoff, k8s.ExistsConditionCheck, k8s.ZookeeperClusterReady)
		if err != nil {
			return err
		}
	} else {
		yaml, err := objects.YAMLManifest()
		if err != nil {
			return err
		}
		fmt.Fprintf(cli.Out(), yaml)
	}

	return nil
}
