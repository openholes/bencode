// Copyright 2025 openHoles Authors
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

package bencode

import (
	"errors"
	"reflect"
)

// Marshal encode any object to bencode data.
func Marshal(v any) ([]byte, error) {
	return marshalValue(reflect.ValueOf(v))
}

// Unmarshal decode bencode data to the target object
func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("must be a non-nil pointer")
	}

	result, err := parseBencode(data)
	if err != nil {
		return err
	}

	return assignValue(rv.Elem(), result)
}
