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

package questionnaire

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
)

type ValidateFuncs map[string]survey.Validator

func GetQuestionsFromStruct(obj interface{}, additionalValidators ValidateFuncs) ([]*survey.Question, error) {
	qs := make([]*survey.Question, 0)

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for k := 0; k < t.NumField(); k++ {
		f := t.Field(k)
		elem := v.Field(k)
		desc := f.Tag.Get("survey.question")
		if desc != "" {
			var def string
			switch elem.Kind() {
			case reflect.Float32:
				def = strconv.Itoa(int(elem.Interface().(float32)))
			case reflect.Int:
				def = strconv.Itoa(int(elem.Interface().(int)))
			case reflect.Int32:
				def = strconv.Itoa(int(elem.Interface().(int32)))
			case reflect.Int64:
				def = strconv.Itoa(int(elem.Interface().(int64)))
			case reflect.String:
				def = elem.Interface().(string)
			default:
				return nil, errors.Errorf("unsupported field type: %s", elem.Type())
			}

			validators := defaultValidators()
			for k, v := range additionalValidators {
				validators[k] = v
			}

			validations := strings.Split(f.Tag.Get("survey.validate"), ",")
			for k, v := range validations {
				validations[k] = strings.TrimSpace(v)
			}

			qs = append(qs, &survey.Question{
				Name:   f.Name,
				Prompt: &survey.Input{Message: desc, Default: def},
				Validate: func(ans interface{}) error {
					for _, validator := range validations {
						var validateFunc survey.Validator
						var ok bool
						if validateFunc, ok = validators[validator]; ok {
							err := validateFunc(ans)
							if err != nil {
								return err
							}
						}
					}

					return nil
				},
			})
		}
	}

	return qs, nil
}

func defaultValidators() map[string]survey.Validator {
	return map[string]survey.Validator{
		"int": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				i, err := strconv.Atoi(s)
				if err != nil {
					return err
				}
				if i <= 0 {
					return errors.New("value must be greater than 0")
				}
			} else {
				return errors.New("invalid input type")
			}
			return nil
		},
		"durationstring": func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				_, err := time.ParseDuration(s)
				return err
			}
			return errors.New("invalid input type")
		},
	}
}
