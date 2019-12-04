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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"

	"emperror.dev/errors"

	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

const qnameCharFmt string = "[A-Za-z0-9]"
const qnameExtCharFmt string = "[-A-Za-z0-9_.]"
const qualifiedNameFmt string = "(" + qnameCharFmt + qnameExtCharFmt + "*)?" + qnameCharFmt

var qualifiedNameRegexp = regexp.MustCompile("^" + qualifiedNameFmt + "$")

type Destination v1alpha3.Destination

func (d Destination) String() string {
	s := d.Host

	if d.Port != nil && d.Port.Number > 0 {
		s = fmt.Sprintf("%s:%d", s, d.Port.Number)
	}

	if d.Subset != nil {
		s = fmt.Sprintf("%s%c(%s)", s, util.Nbsp, *d.Subset)
	}

	if s == "" {
		s = "-"
	}

	return s
}

func ParseDestinations(ds []string) ([]v1alpha3.Destination, error) {
	destinations := make([]v1alpha3.Destination, 0)
	for _, d := range ds {
		destination, err := ParseDestination(d)
		if err != nil {
			return destinations, err
		}
		destinations = append(destinations, destination)
	}

	return destinations, nil
}

func ParseDestination(d string) (v1alpha3.Destination, error) {
	var destination v1alpha3.Destination

	parts := strings.SplitN(d, ":", 3)
	if len(parts) > 0 {
		destination.Host = parts[0]
	}
	if len(parts) > 1 && parts[1] != "" {
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return destination, err
		}
		destination.Port = &v1alpha3.PortSelector{
			Number: uint32(port),
		}
	}
	if len(parts) > 2 {
		destination.Subset = &parts[2]
	}

	if !qualifiedNameRegexp.MatchString(destination.Host) {
		return destination, errors.Errorf("invalid destination host name: %s", destination.Host)
	}

	if destination.Subset != nil && !qualifiedNameRegexp.MatchString(*destination.Subset) {
		return destination, errors.Errorf("invalid destination subset name: %s", *destination.Subset)
	}

	return destination, nil
}
