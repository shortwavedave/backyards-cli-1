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

package nodeexporter

import (
	"emperror.dev/errors"
	"istio.io/operator/pkg/object"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/nodeexporter"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s/resourcemanager"
)

type Manager struct {
	resourcemanager.Manager
	namespace string
}

func NewNodeExporterManager(manager resourcemanager.Manager, namespace string) (*Manager, error) {
	r := &Manager{
		namespace: namespace,
	}

	objects, err := r.getObjects()
	if err != nil {
		return nil, err
	}

	r.Manager = manager
	r.Manager.SetObjects(objects)

	return r, nil
}

func (r *Manager) getObjects() (object.K8sObjects, error) {
	var values ChartValues

	valuesYAML, err := helm.GetDefaultValues(nodeexporter.Chart)
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

	objects, err := helm.Render(nodeexporter.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "prometheus-node-exporter",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: r.namespace,
	}, "prometheus-node-exporter")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}
