// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"errors"
	"reflect"
	"testing"
)

type testMapInterface struct {
	A int
}

func (t *testMapInterface) test() {}

func TestDecodeMap(t *testing.T) {
	type testMap3 struct {
		E int
	}
	type testMap2 struct {
		C int
		D testMap3
	}
	type testMap1 struct {
		A int
		B testMap2
	}
	type testMapKind struct {
		Bool    bool
		BoolPtr *bool
		Int     int
		Float64 float64
		Array   [1][2]int
		// Func      func(int) int
		Interface    interface{}
		InterfacePtr *interface{}
		Map          map[int]map[int]int
		MapPtr       *map[int]map[int]int
		Ptr          *int
		Slice        [][]int
		SlicePtr     *[][]int
		String       string
		Struct       [][]*testMap1
		Error        error
		Default      string
		Rename       string `map:"alt_name"`
		Required     string `map:",required"`
		Omitted      string `map:"-"`
	}

	type testMapI struct {
		A interface{}
		B interface{}
		C interface{ test() }
		D interface{ test() }
		E interface{ test() }
	}

	type nilSlice struct {
		A   [][]*nilSlice
		B   [][]*nilSlice
		C   [][]*nilSlice
		Int int
	}
	type nilArray struct {
		A   [2][2]*nilArray
		B   [2][2]*nilArray
		Int int
	}
	type nilKind1 struct {
		Int int
	}
	type nilKind struct {
		Ptr       *int
		Slice1    [][]*int
		Slice2    [][]*int
		Slice3    [][]*int
		Array1    [2][2]*int
		Array2    [2][2]*int
		Map       map[int]int
		Error     error
		Chan      chan int
		Interface interface{}
		Struct1   *nilKind1
		Struct2   *nilKind1
		Struct3   *nilKind1
		Struct4   nilKind1
		Struct5   nilKind1
	}

	testInt := 1
	testBool := true
	testInterface := interface{}(testInt)
	testMap := map[int]map[int]int{4: {5: 6}}
	testSlicePtr := [][]int{{7}}
	var testInterfaceZero testMapInterface
	var testStruct2Nil *nilKind1
	var testStruct4Zero nilKind1
	tests := []struct {
		name string
		in   map[string]interface{}
		want interface{}
		got  interface{}
	}{
		{
			name: "variation",
			in: map[string]interface{}{
				"Bool":    true,
				"BoolPtr": true,
				"Int":     2,
				"Float64": 0.3,
				"Array":   [1][2]int{{3, 4}},
				// "Func":      func(a int) int { return a },
				"Interface":    5,
				"InterfacePtr": 1,
				"Map":          map[int]map[int]int{3: {2: 1}},
				"MapPtr":       &testMap,
				"Ptr":          &testInt,
				"Slice":        [][]int{{6}},
				"SlicePtr":     &testSlicePtr,
				"String":       "s",
				"Struct": [][]map[string]interface{}{{{
					"A": 7,
					"B": map[string]interface{}{
						"C": 8,
						"D": map[string]interface{}{
							"E": 9,
						},
					},
				}}},
				"Error":    errors.New("e"),
				"alt_name": "a",
				"Required": "r",
				"Omitted":  "-",
			},
			want: testMapKind{
				Bool:    true,
				BoolPtr: &testBool,
				Int:     2,
				Float64: 0.3,
				Array:   [1][2]int{{3, 4}},
				// Func:      func(a int) int { return a },
				Interface:    5,
				InterfacePtr: &testInterface,
				Map:          map[int]map[int]int{3: {2: 1}},
				MapPtr:       &testMap,
				Ptr:          &testInt,
				Slice:        [][]int{{6}},
				SlicePtr:     &testSlicePtr,
				String:       "s",
				Struct: [][]*testMap1{{{
					A: 7,
					B: testMap2{
						C: 8,
						D: testMap3{
							E: 9,
						},
					},
				}}},
				Error:    errors.New("e"),
				Default:  "d",
				Rename:   "a",
				Required: "r",
				Omitted:  "",
			},
			got: &testMapKind{Default: "d"},
		},
		{
			name: "interface",
			in: map[string]interface{}{
				"A": 5,
				"B": testMapI{A: 1},
				"C": &testMapInterface{A: 1},
				"D": &testInterfaceZero,
				"E": &testMapI{},
			},
			want: testMapI{
				A: 5,
				B: testMapI{A: 1},
				C: &testMapInterface{A: 1},
				D: &testInterfaceZero,
				E: nil,
			},
			got: &testMapI{},
		},
		{
			name: "nil slice",
			in: map[string]interface{}{
				"A": [][]map[string]interface{}{{nil, nil}},
				"B": [][]map[string]interface{}{nil, nil},
				"C": [][]map[string]interface{}{{nil, {"A": nil, "Int": 1}}},
			},
			want: nilSlice{
				A: [][]*nilSlice{{nil, nil}},
				B: [][]*nilSlice{nil, nil},
				C: [][]*nilSlice{{nil, &nilSlice{Int: 1}}},
			},
			got: &nilSlice{},
		},
		{
			name: "nil array",
			in: map[string]interface{}{
				"A": [2][2]map[string]interface{}{{nil}},
				"B": [2][2]map[string]interface{}{{{"A": nil, "Int": 1}}},
			},
			want: nilArray{
				A: [2][2]*nilArray{{nil}},
				B: [2][2]*nilArray{{&nilArray{Int: 1}}},
			},
			got: &nilArray{},
		},
		{
			name: "nil variation",
			in: map[string]interface{}{
				"Ptr":       nil,
				"Slice1":    nil,
				"Slice2":    [][]*int{nil},
				"Slice3":    [][]*int{{nil}},
				"Array1":    [2][2]*int{},
				"Array2":    [2][2]*int{{nil}},
				"Map":       nil,
				"Error":     nil,
				"Chan":      nil,
				"Interface": nil,
				"Struct1":   nil,
				"Struct2":   testStruct2Nil,
				"Struct3":   &nilKind1{},
				"Struct4":   nilKind1{},
				"Struct5":   testStruct4Zero,
			},
			want: nilKind{
				Ptr:       nil,
				Slice1:    nil,
				Slice2:    [][]*int{nil},
				Slice3:    [][]*int{{nil}},
				Array1:    [2][2]*int{},
				Array2:    [2][2]*int{{nil}},
				Map:       nil,
				Error:     nil,
				Chan:      nil,
				Interface: nil,
				Struct1:   nil,
				Struct2:   nil,
				Struct3:   &nilKind1{},
				Struct4:   nilKind1{},
				Struct5:   nilKind1{},
			},
			got: &nilKind{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := DecodeMap(tt.in, tt.got, nil)
			if err != nil {
				t.Error(err)
			}
			got := reflect.ValueOf(tt.got).Elem().Interface()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nwant = %+v\ngot  = %+v", tt.want, got)
			}
		})
	}
}
