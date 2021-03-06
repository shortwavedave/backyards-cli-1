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

package servererror

import (
	"errors"

	"github.com/moogar0880/problems"
)

type ErrorCode string

const (
	AuthDisabledErrorCode = ErrorCode("auth-disabled")
)

var (
	ErrAuthDisabled = errors.New("authentication is disabled")
)

type Problem struct {
	problems.DefaultProblem
	// ErrorCode should be safe to match on
	ErrorCode ErrorCode
}
