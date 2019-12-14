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

	"emperror.dev/errors"
	"github.com/MakeNowJust/heredoc"
	"istio.io/operator/pkg/object"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/zookeeper_operator"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

// err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(objects, "ZookeeperCluster"), backoff, k8s.ExistsConditionCheck, k8s.ZookeeperClusterReady)
// if err != nil {
// 	return err
// }

func GetK8sObjects(namespace string) (object.K8sObjects, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(zookeeper_operator.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(zookeeper_operator.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "zookeeper-operator",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	}, "zookeeper-operator")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	namespaceObj, err := k8s.GetNewNamespaceResource(namespace)
	if err != nil {
		return nil, errors.WrapIf(err, "could not render cert-manager namespace object")
	}

	zookeeperObj, err := GetZookeeperCluster(namespace)
	if err != nil {
		return nil, errors.WrapIf(err, "could not render cert-manager namespace object")
	}

	return append(objects, append(zookeeperObj, namespaceObj...)...), nil
}

func GetZookeeperCluster(namespace string) (object.K8sObjects, error) {
	manifest := fmt.Sprintf(heredoc.Doc(`
	apiVersion: zookeeper.pravega.io/v1beta1
	kind: ZookeeperCluster
	metadata:
	  name: example-zookeepercluster
	  namespace: %s
	spec:
	  replicas: 1
	`), namespace)

	return object.ParseK8sObjectsFromYAMLManifest(manifest)
}
