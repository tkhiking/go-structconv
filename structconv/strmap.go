// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	stringMapTagName          = "strmap"
	msgDetailInvalidFieldType = "%v most be %v"
)

type stringMap interface {
	Get(string) (string, bool)
}

type stringMapData struct {
	Data map[string]string
}

func (d *stringMapData) Get(key string) (string, bool) {
	v, ok := d.Data[key]
	return v, ok
}

type getStringMapKeyFn func(*fieldInfo, decodeTagInfo) string

// DecodeStringMap decodes a string map into a struct.
func DecodeStringMap(m map[string]string, v interface{}) error {
	if err := initStruct(v); err != nil {
		return err
	}
	parser := func(f reflect.StructField) (interface{}, error) {
		return parseDecodeTag(f, stringMapTagName)
	}
	info, err := getStructInfo(v, parser)
	if err != nil {
		return err
	}
	var strMap stringMap = &stringMapData{m}
	if err := stringMapToStruct(info, strMap, getStringMapKey); err != nil {
		return err
	}
	return nil
}

func getStringMapKey(info *fieldInfo, tag decodeTagInfo) string {
	var key string
	if tag.Ok && tag.Key != "" {
		key = tag.Key
	} else {
		key = info.Meta.Name
	}
	return key
}

func stringMapToStruct(info *structInfo, m stringMap, fn getStringMapKeyFn) *DecodeError {
	if errs := doStringMapToStruct(info, m, fn); len(errs) > 0 {
		err := &DecodeError{
			Detail: errs,
		}
		return err
	}
	return nil
}

func doStringMapToStruct(info *structInfo, m stringMap, fn getStringMapKeyFn) []*DecodeFieldError {
	var errs []*DecodeFieldError
	for _, inf := range info.Fields {
		if len(inf.Collections) > 0 {
			continue
		}
		if inf.Child != nil {
			childErrs := doStringMapToStruct(inf.Child, m, fn)
			if len(childErrs) > 0 {
				errs = append(errs, childErrs...)
			}
			continue
		}
		tag := inf.Tag.(decodeTagInfo)
		if tag.Omitted {
			continue
		}

		key := fn(inf, tag)
		if val, ok := m.Get(key); ok {
			if err := setStringToField(inf, val); err != nil {
				typ := inf.Value.Type().String()
				msg := fmt.Sprintf(msgDetailInvalidFieldType, key, typ)
				err := &DecodeFieldError{
					Name:     key,
					Value:    val,
					Messages: []string{msg},
				}
				errs = append(errs, err)
			}
		} else if tag.Required {
			err := &DecodeFieldError{
				Name:     key,
				Value:    val,
				Messages: []string{fmt.Sprintf(msgDetailRequired, key)},
			}
			errs = append(errs, err)
		}
	}
	return errs
}

func setStringToField(info *fieldInfo, s string) error {
	fv := info.Value

	switch fv.Type().Kind() {
	case reflect.String:
		fv.SetString(s)
	case reflect.Bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		fv.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(s, 0, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(s, 0, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(v)
	}

	return nil
}
