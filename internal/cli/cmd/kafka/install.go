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

package kafka

import (
	"fmt"
	"os"
	"strings"
	"time"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/kafka"
	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/kafka_operator"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/zookeeper"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

var (
	backoff = wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    50,
	}
)

type installCommand struct {
	cli cli.CLI
}

type InstallOptions struct {
	namespace          string
	zookeeperNamespace string
	istioNamespace     string

	DumpResources bool
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{}
}

func NewInstallCommand(cli cli.CLI, options *InstallOptions) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Install kafka cluster",
		Long: `Installs kafka cluster.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.`,
		Example: `  # Default install.
  backyards kafka install

  # Install kafka cluster into a non-default namespace.
  backyards kafka install -n kafka`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if options.namespace == "" {
				options.namespace = kafkaNamespace
			}

			return c.run(options)
		},
	}

	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", "istio-system", "Namespace of Istio sidecar injector")
	cmd.Flags().BoolVarP(&options.DumpResources, "dump-resources", "d", options.DumpResources, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) installCertManager(options *InstallOptions) error {
	var err error
	var scmd *cobra.Command

	log.Info("installing cert manager component")
	scmdOptions := certmanager.NewInstallOptions()
	if options.DumpResources {
		scmdOptions.DumpResources = true
	}
	scmd = certmanager.NewInstallCommand(c.cli, scmdOptions)
	err = scmd.RunE(scmd, nil)
	if err != nil {
		return errors.WrapIf(err, "error during cert-manager install")
	}

	return nil
}

func (c *installCommand) installZookeeper(options *InstallOptions) error {
	var err error
	var scmd *cobra.Command

	log.Info("installing zookeeper component")
	zkOptions := zookeeper.NewInstallOptions(options.zookeeperNamespace)
	if options.DumpResources {
		zkOptions.DumpResources = true
	}
	scmd = zookeeper.NewInstallCommand(c.cli, zkOptions)
	err = scmd.RunE(scmd, nil)
	if err != nil {
		return errors.WrapIf(err, "error during zookeeper install")
	}

	return nil
}

func (c *installCommand) installKafka(options *InstallOptions) error {
	objects, err := getK8sObjects(options.namespace)
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

		err = k8s.ApplyCRDs(client, c.cli.LabelManager(), crds)
		if err != nil {
			return err
		}

		client, err = c.cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.ApplyResourceObjects(client, c.cli.LabelManager(), objs)
		if err != nil {
			return err
		}

		err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(objs, "KafkaCluster"), backoff, k8s.ExistsConditionCheck, k8s.KafkaClusterReady)
		if err != nil {
			return err
		}
	} else {
		yaml, err := objects.YAMLManifest()
		if err != nil {
			return err
		}
		fmt.Fprintf(c.cli.Out(), yaml)
	}

	return nil
}

func (c *installCommand) run(options *InstallOptions) error {
	err := c.installCertManager(options)
	if err != nil {
		return err
	}

	err = c.installZookeeper(options)
	if err != nil {
		return err
	}

	err = c.installKafka(options)
	if err != nil {
		return err
	}

	return nil
}

func getK8sObjects(namespace string) (object.K8sObjects, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(kafka_operator.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	values.SetDefaults()

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(kafka_operator.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "kafka-operator",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	}, "kafka-operator")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	namespaceObj, err := k8s.GetNewNamespaceResource(namespace)
	if err != nil {
		return nil, errors.WrapIf(err, "could not render cert-manager namespace object")
	}

	kk, err := getKafkaResources()
	if err != nil {
		return nil, err
	}

	objects = append(objects, kk...)

	return append(objects, namespaceObj...), nil
}

func getKafkaResources() ([]*object.K8sObject, error) {
	var err error
	files := []string{}

	dir, err := kafka.Assets.Open(".")
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		dirFiles, err := dir.Readdir(-1)
		if err != nil {
			return nil, err
		}

		for _, file := range dirFiles {
			filename := file.Name()
			if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "tpl") || strings.HasSuffix(filename, "json") {
				files = append(files, "./"+filename)
			}
		}
	}

	contents := []byte{}
	for _, f := range files {
		data, err := helm.ReadIntoBytes(kafka.Assets, f)
		if err != nil {
			return nil, err
		}
		contents = append(contents, data...)
	}

	objs, err := object.ParseK8sObjectsFromYAMLManifest(string(contents))
	if err != nil {
		return nil, errors.WrapIf(err, "could not parse KafkaCluster YAML to K8s object")
	}

	for _, obj := range objs {
		metadata := obj.UnstructuredObject().Object["metadata"].(map[string]interface{})
		metadata["namespace"] = kafkaNamespace
	}

	return objs, nil
}
