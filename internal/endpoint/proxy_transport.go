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

package endpoint

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"emperror.dev/errors"
)

type ReplaceTransport struct {
	PathPrepend string

	http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface
func (t *ReplaceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		message := fmt.Sprintf("Error: '%s'\nTrying to reach: '%v'", err.Error(), req.URL.String())
		resp = &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       ioutil.NopCloser(strings.NewReader(message)),
		}
		return resp, nil
	}

	if redirect := resp.Header.Get("Location"); redirect != "" {
		resp.Header.Set("Location", strings.Replace(redirect, t.PathPrepend, "", -1))
		return resp, nil
	}

	contentType := resp.Header.Get("Content-Type")
	contentType = strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	if contentType != "text/html" {
		return resp, nil
	}

	return t.rewriteResponse(resp)
}

func (t *ReplaceTransport) rewriteHTML(reader io.Reader, writer io.Writer) error {
	var err error

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	b = bytes.ReplaceAll(b, []byte(t.PathPrepend), nil)

	_, err = writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (t *ReplaceTransport) rewriteResponse(resp *http.Response) (*http.Response, error) {
	origBody := resp.Body
	defer origBody.Close()

	newContent := &bytes.Buffer{}
	var reader io.Reader = origBody
	var writer io.Writer = newContent
	encoding := resp.Header.Get("Content-Encoding")
	switch encoding {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return nil, errors.WrapIf(err, "could not make gzip reader")
		}
		gzw := gzip.NewWriter(writer)
		defer gzw.Close()
		writer = gzw
	case "":
	default:
		return resp, nil
	}

	err := t.rewriteHTML(reader, writer)
	if err != nil {
		return resp, err
	}

	resp.Body = ioutil.NopCloser(newContent)
	resp.Header.Del("Content-Length")
	resp.ContentLength = int64(newContent.Len())

	return resp, err
}
