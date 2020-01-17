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

package graphql

import "net/http"

type Request struct {
	query  string
	vars   map[string]interface{}
	header http.Header
}

func (r *Request) Query(query string) {
	r.query = query
}

func (r *Request) GetQuery() string {
	return r.query
}

func (r *Request) GetHeader() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *Request) GetVars() map[string]interface{} {
	return r.vars
}

func (r *Request) Var(key string, value interface{}) {
	if r.vars == nil {
		r.vars = make(map[string]interface{})
	}
	r.vars[key] = value
}
