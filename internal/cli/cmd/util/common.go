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

package util

import (
	"regexp"
	"strings"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	Nbsp            rune   = '\u00A0'
)

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

func ParseK8sResourceID(id string) (types.NamespacedName, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return types.NamespacedName{}, errors.Errorf("invalid resource ID: '%s': format must be <namespace>/<name>", id)
	}

	for _, p := range parts {
		validFormat := dns1123LabelRegexp.MatchString(p)
		if !validFormat {
			return types.NamespacedName{}, errors.Errorf("invalid resource ID: '%s': format must be <namespace>/<name>", id)
		}

	}

	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}
