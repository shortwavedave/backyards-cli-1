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
package k8s

import (
	"time"

	"emperror.dev/errors"
	"istio.io/operator/pkg/object"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/pkg/helm"
	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
)

var (
	backoff = wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    50,
	}
)

func ApplyCRDs(client k8sclient.Client, labelManager LabelManager, crds object.K8sObjects) error {
	crds.Sort(helm.InstallObjectOrder())
	err := ApplyResources(client, labelManager, crds)
	if err != nil {
		return errors.WrapIf(err, "could not apply k8s resources")
	}

	err = WaitForResourcesConditions(client, NamesWithGVKFromK8sObjects(crds), backoff, CRDEstablishedConditionCheck)
	if err != nil {
		return err
	}

	return nil
}

func ApplyResourceObjects(client k8sclient.Client, labelManager LabelManager, objects object.K8sObjects) error {
	objects.Sort(helm.InstallObjectOrder())
	err := ApplyResources(client, labelManager, objects)
	if err != nil {
		return errors.WrapIf(err, "could not apply k8s resources")
	}

	err = WaitForResourcesConditions(client, NamesWithGVKFromK8sObjects(objects, "StatefulSet", "Deployment"), backoff, ExistsConditionCheck, ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}

	return nil
}
