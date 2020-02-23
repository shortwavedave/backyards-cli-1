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
	"context"
	"fmt"
	"strings"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"istio.io/operator/pkg/object"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

type Object interface {
	metav1.Object
	metav1.Type
	schema.ObjectKind
}

type LabelManager interface {
	CheckLabelsBeforeUpdate(actual, desired *unstructured.Unstructured) (bool, error)
	CheckLabelsBeforeCreate(actual *unstructured.Unstructured) (bool, error)
	CheckLabelsBeforeDelete(actual *unstructured.Unstructured) (bool, error)
}

type PostResourceApplyFunc func(k8sclient.Client, Object) error

func ApplyResources(client k8sclient.Client, labelManager LabelManager, objects object.K8sObjects, waitFuncs ...WaitForResourceConditionsFunc) error {
	var err error

	for _, obj := range objects {
		create := true
		actual := obj.UnstructuredObject().DeepCopy()
		desired := obj.UnstructuredObject().DeepCopy()
		desiredCopy := obj.UnstructuredObject().DeepCopy()

		objectName := GetFormattedName(desired)

		if err = client.Get(context.Background(), types.NamespacedName{
			Name:      actual.GetName(),
			Namespace: actual.GetNamespace(),
		}, actual); err == nil {
			create = false
			skip, err := labelManager.CheckLabelsBeforeUpdate(actual, desired)
			if err != nil {
				log.Errorf("%s failed to check labels: %s", objectName, err)
				continue
			}
			if skip {
				log.Warnf("%s skipping resource", objectName)
				continue
			}
			desired.SetResourceVersion(actual.GetResourceVersion())
			patchResult, err := patch.DefaultPatchMaker.Calculate(actual, desired)
			if err != nil {
				log.Error(err, "could not match objects", "object", actual.GetKind())
			} else if patchResult.IsEmpty() {
				log.Infof("%s unchanged", GetFormattedName(actual))
				continue
			}

			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "failed to set last applied annotation", "desired", desired)
			}

			desired = prepareObjectBeforeUpdate(actual, desired)

			err = client.Update(context.Background(), desired)
			if k8serrors.IsConflict(err) || k8serrors.IsInvalid(err) {
				err = client.Delete(context.Background(), desired)
				if err != nil {
					return errors.WrapIfWithDetails(err, "could not delete resource", "name", objectName)
				}
				log.Infof("%s deleted", objectName)
				desired = desiredCopy
				create = true
			} else {
				if err != nil {
					return errors.WrapIfWithDetails(err, "could not update resource", "name", objectName)
				}
				log.Infof("%s configured", objectName)
			}
		}

		if create {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "failed to set last applied annotation", "desired", desired)
			}
			skip, err := labelManager.CheckLabelsBeforeCreate(desired)
			if err != nil {
				log.Errorf("%s failed to check labels: %s", objectName, err)
				continue
			}
			if skip {
				log.Warnf("%s skipping resource", objectName)
				continue
			}
			err = client.Create(context.Background(), desired)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not create resource", "name", objectName)
			}
			log.Infof("%s created", objectName)
		}

		if len(waitFuncs) > 0 {
			for _, fn := range waitFuncs {
				err = fn(client, actual)
				if err != nil {
					log.Error(err)
					continue
				}
			}
		}
	}

	return nil
}

type PostResourceDeleteFunc func(k8sclient.Client, Object) error

func DeleteResources(client k8sclient.Client, labelManager LabelManager, objects object.K8sObjects, waitFuncs ...WaitForResourceConditionsFunc) error {
	var err error

	for _, obj := range objects {
		actual := obj.UnstructuredObject().DeepCopy()
		objectName := GetFormattedName(actual)
		if err = client.Get(context.Background(), types.NamespacedName{
			Name:      actual.GetName(),
			Namespace: actual.GetNamespace(),
		}, actual); err == nil {
			skip, err := labelManager.CheckLabelsBeforeDelete(actual)
			if err != nil {
				log.Errorf("%s failed to check labels: %s", objectName, err)
				continue
			}
			if skip {
				log.Warnf("%s skipping resource", objectName)
				continue
			}
			err = client.Delete(context.Background(), obj.UnstructuredObject())
			if k8serrors.IsNotFound(err) || k8smeta.IsNoMatchError(err) {
				log.Debug(errors.WrapIf(err, "could not delete"))
				continue
			}
			if err != nil {
				log.Error(err)
			}

			deletionTimedOut := false
			if len(waitFuncs) > 0 {
				for _, fn := range waitFuncs {
					err = fn(client, actual)
					if err != nil {
						deletionTimedOut = true
						log.Error(err)
						continue
					}
				}
			}

			if deletionTimedOut {
				log.Errorf("%s deletion timed out", objectName)
			} else {
				log.Infof("%s deleted", objectName)
			}
		} else {
			err = errors.WrapIf(err, "could not delete")
			if k8serrors.IsNotFound(errors.Cause(err)) || k8smeta.IsNoMatchError(errors.Cause(err)) {
				log.Debug(err)
			} else {
				log.Error(err)
			}
		}
	}

	return nil
}

func GetFormattedName(object Object) string {
	var group string
	if object.GroupVersionKind().Group != "" {
		group = "." + object.GroupVersionKind().Group
	}

	namespace := ""
	if object.GetNamespace() != "" {
		namespace = object.GetNamespace() + "/"
	}
	return fmt.Sprintf("%s%s:%s%s", strings.ToLower(object.GetKind()), group, namespace, object.GetName())
}

func prepareObjectBeforeUpdate(actual, desired *unstructured.Unstructured) *unstructured.Unstructured {
	object := desired.DeepCopy()
	if object.GetKind() == "Service" {
		object.Object["spec"].(map[string]interface{})["clusterIP"] = actual.Object["spec"].(map[string]interface{})["clusterIP"]
	}

	return object
}
