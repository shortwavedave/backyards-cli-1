// Copyright © 2020 Banzai Cloud
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

package ale

import (
	"encoding/base64"

	"emperror.dev/errors"
	"github.com/gogo/protobuf/proto"
	pstruct "github.com/golang/protobuf/ptypes/struct"
)

var metadataHeaders = []string{
	"x-by-metadata",
	"x-envoy-peer-metadata",
}

func GetMetadataAttributes(metadata string) (map[string]interface{}, error) {
	decoded, err := base64.StdEncoding.DecodeString(metadata)
	if err != nil {
		return nil, errors.WrapIf(err, "could not base64 decode metadata string")
	}

	var md pstruct.Struct
	err = proto.Unmarshal(decoded, &md)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal metadata proto")
	}

	return getStructValues(&md), nil
}

func getStructValues(md *pstruct.Struct) map[string]interface{} {
	attrs := make(map[string]interface{})

	for k, v := range md.GetFields() {
		attrs[k] = getValue(v)
	}

	return attrs
}

func getListValues(v *pstruct.ListValue) []interface{} {
	values := make([]interface{}, 0)
	for _, value := range v.GetValues() {
		values = append(values, getValue(value))
	}

	return values
}

func getValue(v *pstruct.Value) interface{} {
	switch v.Kind.(type) {
	case *pstruct.Value_StructValue:
		return getStructValues(v.GetStructValue())
	case *pstruct.Value_StringValue:
		return v.GetStringValue()
	case *pstruct.Value_BoolValue:
		return v.GetBoolValue()
	case *pstruct.Value_ListValue:
		return getListValues(v.GetListValue())
	case *pstruct.Value_NumberValue:
		return v.GetNumberValue()
	}

	return nil
}
