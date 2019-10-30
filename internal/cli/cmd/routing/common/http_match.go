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
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"emperror.dev/errors"

	"github.com/banzaicloud/istio-client-go/pkg/common/v1alpha1"
	"github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
)

const defaultMatchType = "exact"

var urlMatcherRegex = regexp.MustCompile(`^((GET|POST|PUT|HEAD|DELETE|PATCH|OPTIONS):)?((https?|ftp)://(-.)?(([^\s\/?\.#-]+.?)+)?(/[^\s]*)?)$`)

type HTTPMatchRequests []v1alpha3.HTTPMatchRequest
type HTTPMatchRequest v1alpha3.HTTPMatchRequest
type StringMatch v1alpha1.StringMatch

func ConvertHTTPMatchRequestsPointers(mrs []*v1alpha3.HTTPMatchRequest) []v1alpha3.HTTPMatchRequest {
	converted := make([]v1alpha3.HTTPMatchRequest, 0)
	for _, mr := range mrs {
		if mr != nil {
			converted = append(converted, *mr)
		}
	}

	return converted
}

func (r HTTPMatchRequests) String() string {
	mrs := make([]string, 0)
	for _, mr := range r {
		mrs = append(mrs, HTTPMatchRequest(mr).String())
	}

	if len(mrs) > 1 {
		return "(" + strings.Join(mrs, ") OR (") + ")"
	}

	if len(mrs) == 1 {
		return mrs[0]
	}

	return "any"
}

func (r StringMatch) String() string {
	if r.Prefix != "" {
		return fmt.Sprintf("%s=%s", "prefix", r.Prefix)
	}
	if r.Suffix != "" {
		return fmt.Sprintf("%s=%s", "suffix", r.Suffix)
	}
	if r.Regex != "" {
		return fmt.Sprintf("%s=%s", "regex", r.Regex)
	}

	return fmt.Sprintf("=%s", r.Exact)
}

func (r HTTPMatchRequest) String() string {
	m := make([]string, 0)

	if r.URI != nil {
		m = append(m, fmt.Sprintf("uri:%s", StringMatch(*r.URI).String()))
	}

	if r.Scheme != nil {
		m = append(m, fmt.Sprintf("scheme:%s", StringMatch(*r.Scheme).String()))
	}

	if r.Method != nil {
		m = append(m, fmt.Sprintf("method:%s", StringMatch(*r.Method).String()))
	}

	if r.Authority != nil {
		m = append(m, fmt.Sprintf("authority:%s", StringMatch(*r.Authority).String()))
	}

	if r.Port != nil {
		m = append(m, fmt.Sprintf("port:%s", strconv.Itoa(int(*r.Port))))
	}

	for k, v := range r.Headers {
		m = append(m, fmt.Sprintf("header:%s:%s", k, StringMatch(v).String()))
	}

	for k, v := range r.QueryParams {
		if v != nil {
			m = append(m, fmt.Sprintf("header:%s:%s", k, StringMatch(*v).String()))
		}
	}

	for k, v := range m {
		m[k] = strings.ReplaceAll(v, ":=", "=")
	}

	return strings.Join(m, " AND ")
}

func ParseHTTPRequestMatches(matchGroups []string) ([]*v1alpha3.HTTPMatchRequest, error) {
	m := make([]*v1alpha3.HTTPMatchRequest, 0)

	for _, mg := range matchGroups {
		rawMatches, err := parseArgs(mg, ',')
		if err != nil {
			return m, err
		}
		match, err := ParseHTTPRequestMatch(rawMatches)
		if err != nil {
			return m, err
		}
		if match != nil {
			m = append(m, match)
		}
	}

	return m, nil
}

func ParseHTTPRequestMatch(matches []string) (*v1alpha3.HTTPMatchRequest, error) {
	var err error
	var m v1alpha3.HTTPMatchRequest

	for _, match := range matches {
		if match == "any" {
			return nil, nil
		}
		parts := strings.SplitN(match, "=", 2)
		if len(parts) < 2 {
			return &m, errors.NewWithDetails("malformed match", "match", match)
		}

		selector := strings.SplitN(parts[0], ":", 3)
		field := selector[0]

		if field == "url" {
			m, err = parseHTTPRequestMatchFromURLMatch(parts[1])
			if err != nil {
				return &m, err
			}
			continue
		}

		matchType := defaultMatchType
		if field == "header" || field == "queryParams" {
			if len(selector) < 2 {
				return nil, errors.NewWithDetails("invalid match", "match", match)
			}
			if len(selector) == 3 {
				matchType = selector[2]
			}
		} else if len(selector) == 2 {
			matchType = selector[1]
		}

		switch matchType {
		case "exact", "prefix", "suffix", "regex":
		default:
			return nil, errors.NewWithDetails("invalid match type", "type", matchType, "match", match)
		}

		switch strings.ToLower(selector[0]) {
		case "scheme":
			m.Scheme = newStringMatch(matchType, parts[1])
		case "method":
			m.Method = newStringMatch(matchType, parts[1])
		case "authority":
			m.Authority = newStringMatch(matchType, parts[1])
		case "uri":
			m.URI = newStringMatch(matchType, parts[1])
		case "port":
			if port, err := strconv.Atoi(parts[1]); err == nil {
				p := uint32(port)
				m.Port = &p
			} else {
				return nil, errors.WrapIfWithDetails(err, "invalid port", "match", match)
			}
		case "header":
			if m.Headers == nil {
				m.Headers = make(map[string]v1alpha1.StringMatch)
			}
			m.Headers[selector[1]] = *newStringMatch(matchType, parts[1])
		case "queryparams":
			if m.QueryParams == nil {
				m.QueryParams = make(map[string]*v1alpha1.StringMatch)
			}
			m.QueryParams[selector[1]] = newStringMatch(matchType, parts[1])
		default:
			return nil, errors.NewWithDetails("invalid match field", "match", match, "field", field)
		}
	}

	return &m, nil
}

func newStringMatch(matchType string, value string) *v1alpha1.StringMatch {
	var m v1alpha1.StringMatch

	switch matchType {
	case "exact":
		m.Exact = value
	case "prefix":
		m.Prefix = value
	case "suffix":
		m.Suffix = value
	case "regex":
		m.Regex = value
	}

	return &m
}

func parseHTTPRequestMatchFromURLMatch(match string) (v1alpha3.HTTPMatchRequest, error) {
	var m v1alpha3.HTTPMatchRequest
	m.Headers = make(map[string]v1alpha1.StringMatch)
	m.QueryParams = make(map[string]*v1alpha1.StringMatch)

	parts := urlMatcherRegex.FindAllStringSubmatch(match, -1)
	if len(parts) == 0 {
		return m, errors.NewWithDetails("malformed url match", "match", match)
	}
	method := parts[0][2]
	URL := parts[0][3]

	u, err := url.Parse(URL)
	if err != nil {
		return m, errors.WrapIfWithDetails(err, "cannot parse url", "match", match)
	}

	if method != "" {
		m.Method = newStringMatch(defaultMatchType, method)
	}

	if u.Hostname() != "" {
		m.Authority = newStringMatch(defaultMatchType, u.Hostname())
	}
	if u.Port() != "" {
		if port, err := strconv.Atoi(u.Port()); err == nil {
			p := uint32(port)
			m.Port = &p
		}
	}
	if u.Path != "" {
		m.URI = newStringMatch(defaultMatchType, u.Path)
	}
	if u.Scheme != "" {
		m.Scheme = newStringMatch(defaultMatchType, u.Scheme)
	}

	for k, v := range u.Query() {
		if len(v) > 0 {
			m.QueryParams[k] = newStringMatch(defaultMatchType, v[0])
		}
	}

	return m, nil
}

const nullStr = rune(0)

func parseArgs(str string, separator rune) ([]string, error) {
	var m []string
	var s string

	str = strings.Trim(str, string(separator)) + string(separator)

	lastQuote := nullStr
	isSeparator := false
	for i, c := range str {
		switch {
		case c == lastQuote:
			lastQuote = nullStr
		case lastQuote != nullStr:
			s += string(c)
		case unicode.In(c, unicode.Quotation_Mark):
			isSeparator = false
			lastQuote = c
		case c == separator:
			if 0 == i || isSeparator {
				continue
			}
			isSeparator = true
			m = append(m, s)
			s = ""
		default:
			isSeparator = false
			s += string(c)
		}
	}

	if lastQuote != nullStr {
		return nil, errors.New("quote did not terminate")
	}

	return m, nil
}
