// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"errors"
	"reflect"
	"strings"
)

var (
	requiredTagValue = "required"
	convTagValue     = "conv"
)

type fieldInfo struct {
	Meta        reflect.StructField
	Value       reflect.Value
	Child       reflect.Value
	ChildOK     bool
	Collections []reflect.Type
}

type decodeTagInfo struct {
	OK       bool
	Key      string
	Required bool
	Omitted  bool
	Conv     bool
}

// checkStructPtr checks the struct pointer.
func checkStructPtr(structPtr interface{}) (reflect.Value, error) {
	pv := reflect.ValueOf(structPtr)
	if pv.Kind() != reflect.Ptr {
		err := errors.New("structconv: structPtr must be a struct pointer")
		return pv, err
	}

	sv := pv.Elem()
	if sv.Kind() != reflect.Struct {
		err := errors.New("structconv: structPtr must be a struct pointer")
		return sv, err
	}
	return sv, nil
}

// walkStructFields walks the structure tree, calling walkFn for each field
// in the tree, including root.
func walkStructFields(s reflect.Value, walkFn func(fieldInfo)) {
	sv := s
	st := sv.Type()
	for i := 0; i < sv.NumField(); i++ {
		fv := sv.Field(i)
		fm := st.Field(i)

		if !fv.CanSet() {
			continue
		}

		child, ok := followStruct(fv, false)
		collections := followStructCollectionsTypes(fv)

		fi := fieldInfo{
			Meta:        fm,
			Value:       fv,
			Child:       child,
			ChildOK:     ok,
			Collections: collections,
		}
		walkFn(fi)
	}
}

func followStruct(fv reflect.Value, init bool) (reflect.Value, bool) {
	// Follow the pointer.
	v := fv
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !init {
				break
			}
			if v.Type().Elem().Kind() != reflect.Struct {
				break
			}
			// Initialize struct pointer.
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		var v reflect.Value
		return v, false
	}
	return v, true
}

func followStructCollectionsTypes(rv reflect.Value) []reflect.Type {
	var collections []reflect.Type
	rt := rv.Type()
	for {
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			collections = append(collections, rt)
		case reflect.Ptr:
			if rt.Elem().Kind() != reflect.Struct {
				return nil
			}
		case reflect.Struct:
			return collections
		default:
			return nil
		}
		rt = rt.Elem()
	}
}

// initStruct initializes the struct pointer.
func initStruct(structPtr interface{}) error {
	sv, err := checkStructPtr(structPtr)
	if err != nil {
		return err
	}
	doInitStruct(sv)
	return nil
}

func doInitStruct(sv reflect.Value) {
	for i := 0; i < sv.NumField(); i++ {
		fv := sv.Field(i)
		if !fv.CanSet() {
			continue
		}
		if cv, ok := followStruct(fv, true); ok {
			doInitStruct(cv)
		}
	}
}

// parseDecodeTag parses the tag for decoding.
func parseDecodeTag(f reflect.StructField, tagName string) (decodeTagInfo, error) {
	var result decodeTagInfo
	tagStr, ok := f.Tag.Lookup(tagName)
	if !ok {
		return result, nil
	}
	result.OK = true

	tags := strings.Split(tagStr, ",")
	for i, v := range tags {
		if i == 0 {
			if v == "-" {
				result.Omitted = true
			} else {
				result.Key = v
			}
			continue
		}
		switch v {
		case requiredTagValue:
			result.Required = true
		case convTagValue:
			result.Conv = true
		}
	}
	return result, nil
}
