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

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

func ParseSidecarEgressBind(bind string) (string, *v1alpha3.Port, error) {
	if bind == "" || strings.HasPrefix(bind, "unix://") {
		return bind, nil, nil
	}

	bindParts := strings.Split(bind, "://")

	if len(bindParts) != 2 {
		return "", nil, errors.New("invalid bind format, please use PROTOCOL://IP:port")
	}

	protocol := v1alpha3.PortProtocol(bindParts[0])

	addressParts := strings.Split(bindParts[1], ":")

	if len(addressParts) != 2 {
		return "", nil, errors.New("invalid bind format, please use PROTOCOL://IP:port")
	}

	portNumber, err := strconv.Atoi(addressParts[1])
	if err != nil {
		return "", nil, errors.New("invalid bind format, port is not an integer")
	}

	return addressParts[0], &v1alpha3.Port{
		Number:   portNumber,
		Protocol: protocol,
	}, nil
}
