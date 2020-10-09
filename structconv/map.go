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

// DecodeMap decodes a map into a struct.
func DecodeMap(m map[string]interface{}, v interface{}) error {
	info, err := getStructInfo(v, parseMapTag)
	if err != nil {
		return err
	}
	if _, decErrs := mapToStruct(m, "map", info); decErrs != nil {
		return &DecodeError{
			Detail: decErrs,
		}
	}
	return nil
}

func parseMapTag(f reflect.StructField) (interface{}, error) {
	return parseDecodeTag(f, mapTagName)
}

func mapToStruct(m interface{}, name string, info *structInfo) (reflect.Value, []*DecodeFieldError) {
	rv := reflect.ValueOf(m)
	keys := map[string]reflect.Value{}
	for _, mk := range rv.MapKeys() {
		keys[mk.String()] = mk
	}
	var errs []*DecodeFieldError

	for fk, f := range info.Fields {
		tag := f.Tag.(decodeTagInfo)
		if tag.Omitted {
			continue
		}

		var mapKeyStr string
		if tag.Ok && tag.Key != "" {
			mapKeyStr = tag.Key
		} else {
			mapKeyStr = f.Meta.Name
		}
		newName := name + "[" + mapKeyStr + "]"

		mk, ok := keys[mapKeyStr]
		if !ok {
			if tag.Required {
				err := &DecodeFieldError{
					Name:     newName,
					Messages: []string{fmt.Sprintf(msgDetailRequired, fk)},
				}
				errs = append(errs, err)
			}
			continue
		}

		mv := rv.MapIndex(mk)
		if isNil(mv) {
			continue
		}

		if mv.Type().Kind() == reflect.Interface {
			mv = mv.Elem()
		}
		fv := f.Value

		switch mv.Type().Kind() {
		case reflect.Map:
			if f.Child != nil {
				_, err := mapToStruct(mv.Interface(), newName, f.Child)
				if len(err) > 0 {
					errs = append(errs, err...)
					continue
				}
				break
			}
			setReflectValue(fv, mv)
		case reflect.Array, reflect.Slice:
			if len(f.Collections) > 0 {
				if err := checkCollections(mv, newName, f, 0); err != nil {
					errs = append(errs, err...)
					continue
				}
				cv, err := makeCollections(mv, newName, f, 0)
				if len(err) > 0 {
					errs = append(errs, err...)
					continue
				}
				fv.Set(cv)
				break
			}
			setReflectValue(fv, mv)
		default:
			setReflectValue(fv, mv)
		}
	}

	return info.Value, errs
}

func setReflectValue(dst, src reflect.Value) {
	if src.Type() == dst.Type() ||
		dst.Type().Kind() == reflect.Interface &&
			src.Type().Implements(dst.Type()) {
		dst.Set(src)
	}
}

func checkCollections(in reflect.Value, name string, info *fieldInfo, depth int) []*DecodeFieldError {
	out := info.Collections[depth]
	var errs []*DecodeFieldError

	if in.Kind() != out.Kind() {
		err := &DecodeFieldError{
			Name:     name,
			Messages: []string{fmt.Sprintf(msgDetailInvalidType, in.Type(), out)},
		}
		return append(errs, err)
	}
	if in.Kind() == reflect.Array && in.Type().Len() != out.Len() {
		err := &DecodeFieldError{
			Name:     name,
			Messages: []string{fmt.Sprintf(msgDetailInvalidType, in.Type(), out)},
		}
		return append(errs, err)
	}
	if depth >= len(info.Collections)-1 {
		return errs
	}

	for i := 0; i < in.Len(); i++ {
		newName := fmt.Sprintf("%s[%d]", name, i)
		if err := checkCollections(in.Index(i), newName, info, depth+1); err != nil {
			errs = append(errs, err...)
		}
	}
	return errs
}

func makeCollections(in reflect.Value, name string, info *fieldInfo, depth int) (reflect.Value, []*DecodeFieldError) {
	if isNil(in) {
		return reflect.Zero(info.Collections[depth]), nil
	}

	var result reflect.Value
	var errs []*DecodeFieldError
	switch in.Type().Kind() {
	case reflect.Array:
		result = reflect.New(info.Collections[depth]).Elem()
		for i := 0; i < in.Len(); i++ {
			newName := fmt.Sprintf("%s[%d]", name, i)
			if depth < len(info.Collections)-1 {
				v, err := makeCollections(in.Index(i), newName, info, depth+1)
				if len(err) > 0 {
					errs = append(errs, err...)
					continue
				}
				result.Index(i).Set(v)
			} else {
				t := info.Collections[depth].Elem()
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
				s := reflect.New(t)
				inf, err := getStructInfo(s.Interface(), parseMapTag)
				if err != nil {
					decErr := &DecodeFieldError{
						Name: newName,
						Messages: []string{
							err.Error(),
						},
					}
					errs = append(errs, decErr)
					continue
				}
				v, decErrs := mapToStruct(in.Index(i).Interface(), newName, inf)
				if len(decErrs) > 0 {
					errs = append(errs, decErrs...)
					continue
				}
				if ptr {
					v = v.Addr()
				}
				result.Index(i).Set(v)
			}
		}
	case reflect.Slice:
		result = reflect.MakeSlice(info.Collections[depth], 0, in.Len())
		for i := 0; i < in.Len(); i++ {
			newName := fmt.Sprintf("%s[%d]", name, i)
			if depth < len(info.Collections)-1 {
				v, err := makeCollections(in.Index(i), newName, info, depth+1)
				if len(err) > 0 {
					errs = append(errs, err...)
					continue
				}
				result = reflect.Append(result, v)
			} else {
				t := info.Collections[depth].Elem()
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
				s := reflect.New(t)
				inf, err := getStructInfo(s.Interface(), parseMapTag)
				if err != nil {
					decErr := &DecodeFieldError{
						Name: newName,
						Messages: []string{
							err.Error(),
						},
					}
					errs = append(errs, decErr)
					continue
				}
				v, decErrs := mapToStruct(in.Index(i).Interface(), newName, inf)
				if len(decErrs) > 0 {
					errs = append(errs, decErrs...)
					continue
				}
				if ptr {
					v = v.Addr()
				}
				result = reflect.Append(result, v)
			}
		}
	}

	return result, errs
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
