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
	"fmt"
	"os"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

type labelManager struct {
	skipAll     bool
	manageAll   bool
	deleteAll   bool
	interactive bool
	version     string
}

type response string

const (
	CLIVersionLabel = "backyards.banzaicloud.io/cli-version"

	Skip      response = "Skip this resource"
	SkipAll   response = "Skip all"
	Manage    response = "Manage this resource from now on"
	ManageAll response = "Manage all"
	Delete    response = "Delete"
	DeleteAll response = "Delete all"
	Quit      response = "Quit"
)

func NewLabelManager(interactive bool, version string) k8s.LabelManager {
	return &labelManager{
		interactive: interactive,
		version:     version,
	}
}

func (lm *labelManager) CheckLabelsBeforeCreate(desired *unstructured.Unstructured) (bool, error) {
	lm.setDesiredLabels(desired)
	return false, nil
}

func (lm *labelManager) CheckLabelsBeforeDelete(actual *unstructured.Unstructured) (bool, error) {
	labels := actual.GetLabels()
	if _, ok := labels[CLIVersionLabel]; !ok {
		if lm.deleteAll {
			return false, nil
		}
		if !lm.skipAll {
			if lm.interactive {
				var r string
				prompt := &survey.Select{
					Message: fmt.Sprintf("Existing resource %s is not managed by us",
						k8s.GetFormattedName(actual)),
					Options: []string{string(Skip), string(SkipAll), string(Delete), string(DeleteAll), string(Quit)},
				}
				err := survey.AskOne(prompt, &r)
				if err != nil {
					return true, err
				}
				switch response(r) {
				default:
					return true, errors.Errorf("invalid response %s", response(r))
				case Quit:
					os.Exit(0)
				case Skip:
					return true, nil
				case SkipAll:
					lm.skipAll = true
					return true, nil
				case Delete:
					lm.deleteAll = true
					return false, nil
				case DeleteAll:
					return false, nil
				}
			}
		}
	}
	return false, nil
}

func (lm *labelManager) CheckLabelsBeforeUpdate(actual, desired *unstructured.Unstructured) (bool, error) {
	labels := actual.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	if _, ok := labels[CLIVersionLabel]; !ok {
		if lm.manageAll {
			lm.setDesiredLabels(desired)
			return false, nil
		}
		if !lm.skipAll {
			if lm.interactive {
				var r string
				prompt := &survey.Select{
					Message: fmt.Sprintf("Existing resource %s is not yet managed by us",
						k8s.GetFormattedName(actual)),
					Options: []string{string(Skip), string(SkipAll), string(Manage), string(ManageAll), string(Quit)},
				}
				err := survey.AskOne(prompt, &r)
				if err != nil {
					return true, err
				}
				switch response(r) {
				default:
					return true, errors.Errorf("invalid response %s", response(r))
				case Quit:
					os.Exit(0)
				case Skip:
					return true, nil
				case SkipAll:
					lm.skipAll = true
					return true, nil
				case ManageAll:
					lm.manageAll = true
					lm.setDesiredLabels(desired)
					return false, nil
				case Manage:
					lm.setDesiredLabels(desired)
					return false, nil
				}
			}
		}
		// skip everything otherwise
		return true, nil
	}
	// update existing labels with the current version
	lm.setDesiredLabels(desired)
	return false, nil
}

func (lm *labelManager) setDesiredLabels(desired *unstructured.Unstructured) {
	labels := desired.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	if _, ok := labels[CLIVersionLabel]; !ok {
		labels[CLIVersionLabel] = lm.version
	}
	desired.SetLabels(labels)
}
