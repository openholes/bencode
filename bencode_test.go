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
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		args    any
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "int 1",
			args:    1,
			want:    []byte("i1e"),
			wantErr: assert.NoError,
		},
		{
			name:    "int -1",
			args:    -1,
			want:    []byte("i-1e"),
			wantErr: assert.NoError,
		},
		{
			name:    "int 0",
			args:    0,
			want:    []byte("i0e"),
			wantErr: assert.NoError,
		},
		{
			name:    "int -0",
			args:    -0,
			want:    []byte("i0e"),
			wantErr: assert.NoError,
		},
		{
			name:    "int -65535",
			args:    -65535,
			want:    []byte("i-65535e"),
			wantErr: assert.NoError,
		},
		{
			name:    "string 1",
			args:    "1",
			want:    []byte("1:1"),
			wantErr: assert.NoError,
		},
		{
			name:    "string spam",
			args:    "spam",
			want:    []byte("4:spam"),
			wantErr: assert.NoError,
		},
		{
			name:    "list hello,world",
			args:    []string{"hello", "world"},
			want:    []byte("l5:hello5:worlde"),
			wantErr: assert.NoError,
		},
		{
			name:    "dict t1=v1,t2=2",
			args:    map[string]any{"t1": "v1", "t2": 2},
			want:    []byte("d2:t12:v12:t2i2ee"),
			wantErr: assert.NoError,
		},
		{
			name: "struct test1",
			args: struct {
				Announce string `bencode:"announce"`
				Files    []struct {
					Name       string `bencode:"name"`
					Size       int    `bencode:"size"`
					FloatValue float64
				} `bencode:"files"`
				Created int64 `bencode:"-"`
			}{
				Announce: "https://tracker",
				Files: []struct {
					Name       string `bencode:"name"`
					Size       int    `bencode:"size"`
					FloatValue float64
				}{
					{Name: "file1.txt", Size: 1024, FloatValue: 1.0},
				},
				Created: 12345,
			},
			want:    []byte(`d8:announce15:https://tracker5:filesld4:name9:file1.txt4:sizei1024eeee`),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.args)
			if !tt.wantErr(t, err, fmt.Sprintf("Marshal(%v)", tt.args)) {
				return
			}
			assert.Equalf(t, string(tt.want), string(got), "Marshal(%v)", tt.args)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		data []byte
		v    any
	}
	tests := []struct {
		name      string
		args      args
		wantErr   assert.ErrorAssertionFunc
		wantValue func(v any) bool
	}{
		{
			name:    "int 1",
			args:    args{data: []byte("i1e"), v: new(int)},
			wantErr: assert.NoError,
			wantValue: func(v any) bool {
				vv, ok := v.(*int)
				if !ok {
					return false
				}
				return reflect.DeepEqual(*vv, 1)
			},
		},
		{
			name: "string 1",
			args: args{data: []byte("1:1"), v: new(string)},
			wantValue: func(v any) bool {
				vv, ok := v.(*string)
				if !ok {
					return false
				}
				return reflect.DeepEqual(*vv, "1")
			},
			wantErr: assert.NoError,
		},
		{
			name: "list hello,world",
			args: args{data: []byte("l5:hello5:worlde"), v: new([]string)},
			wantValue: func(v any) bool {
				vv, ok := v.(*[]string)
				if !ok {
					return false
				}
				return reflect.DeepEqual(*vv, []string{"hello", "world"})
			},
			wantErr: assert.NoError,
		},
		{
			name: "dict t1=v1,t2=2",
			args: args{data: []byte("d2:t12:v12:t2i2ee"), v: new(map[string]any)},
			wantValue: func(v any) bool {
				vv, ok := v.(*map[string]any)
				if !ok {
					return false
				}
				return reflect.DeepEqual(*vv, map[string]any{"t1": "v1", "t2": int64(2)})
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, Unmarshal(tt.args.data, tt.args.v), fmt.Sprintf("Unmarshal(%v, %v)", tt.args.data, tt.args.v))
			assert.True(t, tt.wantValue(tt.args.v))
		})
	}
}
