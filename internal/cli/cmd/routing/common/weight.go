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

package common

import (
	"strconv"
	"strings"

	"emperror.dev/errors"
)

func ParseWeights(weights string) ([]int, error) {
	ws := make([]int, 0)

	sum := 0
	for _, w := range strings.Split(weights, "/") {
		i, err := strconv.Atoi(w)
		if err != nil {
			return nil, err
		}
		sum += i
		ws = append(ws, i)
	}

	if sum != 100 {
		return nil, errors.New("sum of the weight must be 100")
	}

	return ws, nil
}
