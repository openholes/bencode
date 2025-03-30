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
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
)

func marshalValue(v reflect.Value) ([]byte, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(fmt.Sprintf("i%de", v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return []byte(fmt.Sprintf("i%de", v.Uint())), nil
	case reflect.String:
		s := v.String()
		return []byte(fmt.Sprintf("%d:%s", len(s), s)), nil
	case reflect.Slice, reflect.Array:
		var buf bytes.Buffer
		buf.WriteByte('l')
		for i := 0; i < v.Len(); i++ {
			elem, err := marshalValue(v.Index(i))
			if err != nil {
				continue
			}
			buf.Write(elem)
		}
		buf.WriteByte('e')
		return buf.Bytes(), nil
	case reflect.Struct:
		fields := make(map[string]reflect.Value)
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			tag := field.Tag.Get("bencode")
			if tag == "-" {
				continue
			}
			key := field.Name
			if tag != "" {
				key = tag
			}
			fields[key] = v.Field(i)
		}

		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var buf bytes.Buffer
		buf.WriteByte('d')
		for _, k := range keys {
			keyBytes := []byte(fmt.Sprintf("%d:%s", len(k), k))
			valBytes, err := marshalValue(fields[k])
			if err != nil {
				continue
			}
			buf.Write(keyBytes)
			buf.Write(valBytes)
		}
		buf.WriteByte('e')
		return buf.Bytes(), nil
	case reflect.Map:
		var buf bytes.Buffer
		buf.WriteByte('d')

		keys := v.MapKeys()
		strKeys := make([]string, 0, len(keys))
		for _, k := range keys {
			if k.Kind() == reflect.String {
				strKeys = append(strKeys, k.String())
			}
		}
		sort.Strings(strKeys)

		for _, sk := range strKeys {
			k := reflect.ValueOf(sk)
			val := v.MapIndex(k)
			keyBytes, _ := marshalValue(k)
			valBytes, err := marshalValue(val)
			if err != nil {
				continue
			}
			buf.Write(keyBytes)
			buf.Write(valBytes)
		}
		buf.WriteByte('e')
		return buf.Bytes(), nil
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return marshalValue(v.Elem())
	default:
		return nil, errors.New("unsupported type")
	}
}
