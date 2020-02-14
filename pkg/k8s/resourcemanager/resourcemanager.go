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

package resourcemanager

import (
	"time"

	"emperror.dev/errors"
	"istio.io/operator/pkg/object"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
)

type Manager interface {
	SetObjects(objects object.K8sObjects)
	Install() Manager
	Uninstall() Manager
	Do() error
	Resources() object.K8sObjects
	YAML() (string, error)
}

type Action string

const (
	InstallAction   Action = "install"
	UninstallAction Action = "uninstall"
)

type ResourcesManager struct {
	objects      object.K8sObjects
	client       k8sclient.Client
	labelManager k8s.LabelManager

	action Action
}

func New(client k8sclient.Client, labelManager k8s.LabelManager) *ResourcesManager {
	return &ResourcesManager{
		client:       client,
		labelManager: labelManager,
	}
}

func (r *ResourcesManager) SetObjects(objects object.K8sObjects) {
	r.objects = objects
}

func (r *ResourcesManager) Install() Manager {
	r.action = InstallAction
	r.objects.Sort(helm.InstallObjectOrder())

	return r
}

func (r *ResourcesManager) Uninstall() Manager {
	r.action = UninstallAction
	r.objects.Sort(helm.UninstallObjectOrder())

	return r
}

func (r *ResourcesManager) Do() error {
	switch r.action {
	case InstallAction:
		return r.install()
	case UninstallAction:
		return r.uninstall()
	}

	return errors.New("unknown action")
}

func (r *ResourcesManager) Resources() object.K8sObjects {
	return r.objects
}

func (r *ResourcesManager) YAML() (string, error) {
	return r.objects.YAMLManifest()
}

func (r *ResourcesManager) install() error {
	var err error

	err = k8s.ApplyResources(r.client, r.labelManager, r.objects)
	if err != nil {
		return err
	}

	err = k8s.WaitForResourcesConditions(r.client, k8s.NamesWithGVKFromK8sObjects(r.objects), wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    24,
	}, k8s.ExistsConditionCheck, k8s.ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResourcesManager) uninstall() error {
	err := k8s.DeleteResources(r.client, r.labelManager, r.objects, k8s.WaitForResourceConditions(wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    24,
	}, k8s.NonExistsConditionCheck))
	if err != nil {
		return errors.WrapIf(err, "could not delete k8s resources")
	}

	return nil
}
