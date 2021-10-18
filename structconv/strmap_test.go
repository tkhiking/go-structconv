// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"reflect"
	"testing"
)

func TestDecodeStringMap(t *testing.T) {
	type testNestedStringMap1 struct {
		N1 int
	}
	type testNestedStringMap2 struct {
		N2 int
	}
	type testStringMap struct {
		String   string
		Bool     bool
		Int      int
		Float64  float64
		Default  string
		Rename   string `strmap:"alt_key"`
		Required string `strmap:",required"`
		Omitted  string `strmap:"-"`
		Nest11   testNestedStringMap1
		Nest12   *testNestedStringMap1
		Nest2    [][][]*testNestedStringMap2
	}

	tests := []struct {
		name string
		in   map[string]string
		want testStringMap
	}{
		{
			"normal",
			map[string]string{
				"String":   "str",
				"Bool":     "true",
				"Int":      "1",
				"Float64":  "0.3",
				"Default":  "d",
				"alt_key":  "alt",
				"Required": "r",
				"Omitted":  "-",
				"N1":       "1",
				"N2":       "2",
			},
			testStringMap{
				"str",
				true,
				1,
				0.3,
				"d",
				"alt",
				"r",
				"",
				testNestedStringMap1{N1: 1},
				&testNestedStringMap1{N1: 1},
				nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got testStringMap
			err := DecodeStringMap(tt.in, &got, nil)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nwant = %+v\ngot  = %+v", tt.want, got)
			}
		})
	}
}
