// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

// Package structconv is a converter between struct and other data.
package structconv

import (
	"fmt"
	"reflect"
)

const (
	msgDetailInvalidType = "invalid type: in=%v, out=%v"
	mapTagName           = "map"
)

type DecodeMapOptions struct {
	TagName string
	TagOnly bool
}

// DecodeMap decodes a map into a struct.
func DecodeMap(m map[string]interface{}, v interface{}, o *DecodeMapOptions) error {
	o = initDecodeMapOptions(o)
	s, err := checkStructPtr(v)
	if err != nil {
		return err
	}
	if decErrs := mapToStruct("map", m, s, *o); len(decErrs) > 0 {
		return &DecodeError{
			Detail: decErrs,
		}
	}
	return nil
}

func initDecodeMapOptions(o *DecodeMapOptions) *DecodeMapOptions {
	if o == nil {
		o = &DecodeMapOptions{
			TagName: mapTagName,
			TagOnly: false,
		}
	}
	if o.TagName == "" {
		o.TagName = mapTagName
	}
	return o
}

func mapToStruct(name string, m interface{}, s reflect.Value, o DecodeMapOptions) []*DecodeFieldError {
	rv := reflect.ValueOf(m)
	var decErrs []*DecodeFieldError

	walkStructFields(s, func(f fieldInfo) {
		fm := f.Meta
		fk := fm.Name

		tag, err := parseDecodeTag(fm, o.TagName)
		if err != nil {
			decErr := &DecodeFieldError{
				Name: name + "[" + fk + "]",
				Messages: []string{
					err.Error(),
				},
			}
			decErrs = append(decErrs, decErr)
			return
		}
		if o.TagOnly && !f.ChildOK && !tag.OK {
			return
		}
		if tag.Omitted {
			return
		}

		var mapKeyStr string
		if tag.OK && tag.Key != "" {
			mapKeyStr = tag.Key
		} else {
			mapKeyStr = fk
		}
		newName := name + "[" + mapKeyStr + "]"

		mv := rv.MapIndex(reflect.ValueOf(mapKeyStr))
		if !mv.IsValid() {
			if tag.Required {
				decErr := &DecodeFieldError{
					Name:     newName,
					Messages: []string{fmt.Sprintf(msgDetailRequired, fk)},
				}
				decErrs = append(decErrs, decErr)
			}
			return
		}

		doMapToStruct(newName, mv, f, tag, o)
	})

	return decErrs
}

func doMapToStruct(name string, mv reflect.Value, fi fieldInfo, tag decodeTagInfo, o DecodeMapOptions) []*DecodeFieldError {
	if isNil(mv) {
		return nil
	}
	if mv.Type().Kind() == reflect.Interface {
		mv = mv.Elem()
	}
	if tag.Conv {
		mv = mv.Convert(fi.Meta.Type)
	}

	switch mv.Type().Kind() {
	case reflect.Map:
		if !fi.ChildOK {
			setReflectValue(fi.Value, mv)
			break
		}
		if e := mapToStruct(name, mv.Interface(), fi.Child, o); len(e) > 0 {
			return e
		}
	case reflect.Array, reflect.Slice:
		if len(fi.Collections) == 0 {
			setReflectValue(fi.Value, mv)
			break
		}
		if e := checkCollections(name, mv, fi.Collections); e != nil {
			return e
		}
		cv, e := makeCollections(name, mv, fi.Collections, o)
		if len(e) > 0 {
			return e
		}
		fi.Value.Set(cv)
	default:
		setReflectValue(fi.Value, mv)
	}
	return nil
}

func setReflectValue(dst, src reflect.Value) {
	if src.Type() == dst.Type() ||
		dst.Type().Kind() == reflect.Interface &&
			src.Type().Implements(dst.Type()) {
		dst.Set(src)
	}
	if dst.Type().Kind() == reflect.Ptr && src.Type().Kind() != reflect.Ptr {
		dstET := dst.Type().Elem()
		eqType := src.Type() == dstET
		impl := dstET.Kind() == reflect.Interface && src.Type().Implements(dstET)
		if (eqType || impl) && !isNil(src) {
			rv := reflect.New(dstET)
			rv.Elem().Set(src)
			dst.Set(rv)
			// dst.Set(src.Addr())
		}
	}
}

func checkCollections(name string, in reflect.Value, out []reflect.Type) []*DecodeFieldError {
	if len(out) == 0 {
		return nil
	}
	var decErrs []*DecodeFieldError

	if in.Kind() != out[0].Kind() {
		msg := fmt.Sprintf(msgDetailInvalidType, in.Type(), out[0])
		decErr := &DecodeFieldError{
			Name:     name,
			Messages: []string{msg},
		}
		return append(decErrs, decErr)
	}
	if in.Kind() == reflect.Array && in.Type().Len() != out[0].Len() {
		msg := fmt.Sprintf(msgDetailInvalidType, in.Type(), out[0])
		decErr := &DecodeFieldError{
			Name:     name,
			Messages: []string{msg},
		}
		return append(decErrs, decErr)
	}
	if len(out) == 1 {
		return nil
	}

	for i := 0; i < in.Len(); i++ {
		newName := name + "[" + fmt.Sprint(i) + "]"
		if e := checkCollections(newName, in.Index(i), out[1:]); len(e) > 0 {
			decErrs = append(decErrs, e...)
		}
	}
	return decErrs
}

func makeCollections(name string, in reflect.Value, out []reflect.Type, o DecodeMapOptions) (reflect.Value, []*DecodeFieldError) {
	if len(out) == 0 {
		var v reflect.Value
		return v, []*DecodeFieldError{{
			Name:     name,
			Messages: []string{"internal error: out is empty"},
		}}
	}
	if isNil(in) {
		return reflect.Zero(out[0]), nil
	}

	var result reflect.Value
	var decErrs []*DecodeFieldError
	switch in.Type().Kind() {
	case reflect.Array:
		if len(out) > 1 {
			result = reflect.New(out[0]).Elem()
			for i := 0; i < in.Len(); i++ {
				newName := name + "[" + fmt.Sprint(i) + "]"
				v, e := makeCollections(newName, in.Index(i), out[1:], o)
				if len(e) > 0 {
					decErrs = append(decErrs, e...)
					continue
				}
				result.Index(i).Set(v)
			}
		} else {
			var e []*DecodeFieldError
			result, e = makeArrayStruct(name, in, out[0], o)
			if len(e) > 0 {
				decErrs = append(decErrs, e...)
			}
		}
	case reflect.Slice:
		if len(out) > 1 {
			result = reflect.MakeSlice(out[0], 0, in.Len())
			for i := 0; i < in.Len(); i++ {
				newName := name + "[" + fmt.Sprint(i) + "]"
				v, e := makeCollections(newName, in.Index(i), out[1:], o)
				if len(e) > 0 {
					decErrs = append(decErrs, e...)
					continue
				}
				result = reflect.Append(result, v)
			}
		} else {
			var e []*DecodeFieldError
			result, e = makeSliceStruct(name, in, out[0], o)
			if len(e) > 0 {
				decErrs = append(decErrs, e...)
			}
		}
	}
	return result, decErrs
}

func makeArrayStruct(name string, in reflect.Value, out reflect.Type, o DecodeMapOptions) (reflect.Value, []*DecodeFieldError) {
	result := reflect.New(out).Elem()
	var decErrs []*DecodeFieldError
	for i := 0; i < in.Len(); i++ {
		newName := name + "[" + fmt.Sprint(i) + "]"
		t := out.Elem()
		if isNil(in.Index(i)) {
			v := reflect.Zero(t)
			result.Index(i).Set(v)
			continue
		}
		ptr := false
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
			ptr = true
		}
		pv := reflect.New(t)
		sv := pv.Elem()
		e := mapToStruct(newName, in.Index(i).Interface(), sv, o)
		if len(e) > 0 {
			decErrs = append(decErrs, e...)
			continue
		}
		v := sv
		if ptr {
			v = v.Addr()
		}
		result.Index(i).Set(v)
	}
	return result, decErrs
}

func makeSliceStruct(name string, in reflect.Value, out reflect.Type, o DecodeMapOptions) (reflect.Value, []*DecodeFieldError) {
	result := reflect.MakeSlice(out, 0, in.Len())
	var decErrs []*DecodeFieldError
	for i := 0; i < in.Len(); i++ {
		newName := name + "[" + fmt.Sprint(i) + "]"
		t := out.Elem()
		if isNil(in.Index(i)) {
			v := reflect.Zero(t)
			result = reflect.Append(result, v)
			continue
		}
		ptr := false
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
			ptr = true
		}
		pv := reflect.New(t)
		sv := pv.Elem()
		e := mapToStruct(newName, in.Index(i).Interface(), sv, o)
		if len(e) > 0 {
			decErrs = append(decErrs, e...)
			continue
		}
		v := sv
		if ptr {
			v = v.Addr()
		}
		result = reflect.Append(result, v)
	}
	return result, decErrs
}

func isNil(v reflect.Value) bool {
	// if v.Kind() != reflect.Func && !v.IsValid() || v.IsZero() {
	// 	return true
	// }
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if v.IsNil() {
			return true
		}
	}
	return false
}
