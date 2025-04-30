// Copyright 2025 opencave Authors
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
	"strconv"
)

func parseBencode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	switch data[0] {
	case 'i':
		end := bytes.IndexByte(data, 'e')
		if end == -1 {
			return nil, errors.New("invalid integer")
		}
		num, err := strconv.ParseInt(string(data[1:end]), 10, 64)
		if err != nil {
			return nil, err
		}
		return num, nil
	case 'l':
		var list []interface{}
		data = data[1:]
		for len(data) > 0 && data[0] != 'e' {
			item, rest, err := parseElement(data)
			if err != nil {
				return nil, err
			}
			list = append(list, item)
			data = rest
		}
		return list, nil
	case 'd':
		dict := make(map[string]interface{})
		data = data[1:]
		for len(data) > 0 && data[0] != 'e' {
			key, rest, err := parseString(data)
			if err != nil {
				return nil, err
			}
			val, rest, err := parseElement(rest)
			if err != nil {
				return nil, err
			}
			dict[key] = val
			data = rest
		}
		return dict, nil
	default:
		if data[0] >= '0' && data[0] <= '9' {
			str, _, err := parseString(data)
			return str, err
		}
		return nil, errors.New("invalid format")
	}
}

func parseElement(data []byte) (interface{}, []byte, error) {
	switch data[0] {
	case 'i':
		end := bytes.IndexByte(data, 'e')
		if end == -1 {
			return nil, nil, errors.New("invalid integer")
		}
		num, err := strconv.ParseInt(string(data[1:end]), 10, 64)
		return num, data[end+1:], err
	case 'l':
		list, rest, err := parseList(data)
		return list, rest, err
	case 'd':
		dict, rest, err := parseDict(data)
		return dict, rest, err
	default:
		if data[0] >= '0' && data[0] <= '9' {
			str, rest, err := parseString(data)
			return str, rest, err
		}
		return nil, nil, errors.New("invalid element")
	}
}

func parseList(data []byte) ([]interface{}, []byte, error) {
	data = data[1:]
	var list []interface{}
	for len(data) > 0 && data[0] != 'e' {
		item, rest, err := parseElement(data)
		if err != nil {
			return nil, nil, err
		}
		list = append(list, item)
		data = rest
	}
	return list, data[1:], nil
}

func parseDict(data []byte) (map[string]interface{}, []byte, error) {
	data = data[1:]
	dict := make(map[string]interface{})
	for len(data) > 0 && data[0] != 'e' {
		key, rest, err := parseString(data)
		if err != nil {
			return nil, nil, err
		}
		val, rest, err := parseElement(rest)
		if err != nil {
			return nil, nil, err
		}
		dict[key] = val
		data = rest
	}
	return dict, data[1:], nil
}

func parseString(data []byte) (string, []byte, error) {
	colon := bytes.IndexByte(data, ':')
	if colon == -1 {
		return "", nil, errors.New("invalid string format")
	}

	length, err := strconv.Atoi(string(data[:colon]))
	if err != nil {
		return "", nil, err
	}

	end := colon + 1 + length
	if end > len(data) {
		return "", nil, errors.New("string length exceeds data")
	}

	return string(data[colon+1 : end]), data[end:], nil
}

func assignValue(dst reflect.Value, src interface{}) error {
	switch dst.Kind() {
	case reflect.Struct:
		srcMap, ok := src.(map[string]interface{})
		if !ok {
			return errors.New("cannot assign non-map to struct")
		}

		t := dst.Type()
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

			if val, ok := srcMap[key]; ok {
				fieldValue := dst.Field(i)
				if err := assignValue(fieldValue, val); err != nil {
					return err
				}
			}
		}
		return nil

	case reflect.Map:
		srcMap, ok := src.(map[string]interface{})
		if !ok {
			return errors.New("cannot assign non-map to map")
		}

		if dst.Type().Key().Kind() != reflect.String {
			return errors.New("map key must be string")
		}

		if dst.IsNil() {
			dst.Set(reflect.MakeMap(dst.Type()))
		}

		elemType := dst.Type().Elem()
		for k, v := range srcMap {
			key := reflect.ValueOf(k)
			elem := reflect.New(elemType).Elem()
			if err := assignValue(elem, v); err != nil {
				return err
			}
			dst.SetMapIndex(key, elem)
		}
		return nil

	case reflect.Slice:
		srcSlice, ok := src.([]interface{})
		if !ok {
			return errors.New("cannot assign non-slice to slice")
		}

		slice := reflect.MakeSlice(dst.Type(), len(srcSlice), len(srcSlice))
		for i, item := range srcSlice {
			if err := assignValue(slice.Index(i), item); err != nil {
				return err
			}
		}
		dst.Set(slice)
		return nil

	case reflect.Ptr:
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assignValue(dst.Elem(), src)

	case reflect.Interface:
		dst.Set(reflect.ValueOf(src))
		return nil

	default:
		srcValue := reflect.ValueOf(src)
		if srcValue.Type().ConvertibleTo(dst.Type()) {
			dst.Set(srcValue.Convert(dst.Type()))
			return nil
		}
		return fmt.Errorf("cannot convert %T to %v", src, dst.Type())
	}
}
