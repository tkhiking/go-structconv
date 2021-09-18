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

type decodeTagParser func(reflect.StructField) (decodeTagInfo, error)

type stringMapToStructParams struct {
	Struct    reflect.Value
	StringMap stringMap
	TagParser decodeTagParser
	KeyParser stringMapKeyFn
}

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

type stringMapKeyFn func(fieldInfo, decodeTagInfo) string

// DecodeStringMap decodes a string map into a struct.
func DecodeStringMap(m map[string]string, v interface{}) error {
	s, err := checkStructPtr(v)
	if err != nil {
		return err
	}
	if err := initStruct(v); err != nil {
		return err
	}
	parser := func(f reflect.StructField) (decodeTagInfo, error) {
		return parseDecodeTag(f, stringMapTagName)
	}
	var strMap stringMap = &stringMapData{m}
	params := stringMapToStructParams{
		Struct:    s,
		StringMap: strMap,
		TagParser: parser,
		KeyParser: getStringMapKey,
	}
	if err := stringMapToStruct(params); err != nil {
		return err
	}
	return nil
}

func getStringMapKey(info fieldInfo, tag decodeTagInfo) string {
	var key string
	if tag.OK && tag.Key != "" {
		key = tag.Key
	} else {
		key = info.Meta.Name
	}
	return key
}

func stringMapToStruct(params stringMapToStructParams) *DecodeError {
	if errs := doStringMapToStruct(params); len(errs) > 0 {
		err := &DecodeError{
			Detail: errs,
		}
		return err
	}
	return nil
}

func doStringMapToStruct(params stringMapToStructParams) []*DecodeFieldError {
	var errs []*DecodeFieldError
	walkStructFields(params.Struct, func(inf fieldInfo) {
		if len(inf.Collections) > 0 {
			return
		}
		if inf.ChildOK {
			p := stringMapToStructParams{
				Struct:    inf.Child,
				StringMap: params.StringMap,
				TagParser: params.TagParser,
				KeyParser: params.KeyParser,
			}
			childErrs := doStringMapToStruct(p)
			if len(childErrs) > 0 {
				errs = append(errs, childErrs...)
			}
			return
		}
		tag, err := params.TagParser(inf.Meta)
		if err != nil {
			decErr := &DecodeFieldError{
				Name: inf.Meta.Name,
				Messages: []string{
					err.Error(),
				},
			}
			errs = append(errs, decErr)
			return
		}
		if tag.Omitted {
			return
		}

		key := params.KeyParser(inf, tag)
		if val, ok := params.StringMap.Get(key); ok {
			if err := convertStringToField(inf.Value, val); err != nil {
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
	})
	return errs
}

func convertStringToField(rv reflect.Value, s string) error {
	switch rv.Type().Kind() {
	case reflect.String:
		rv.SetString(s)
	case reflect.Bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		rv.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(s, 0, rv.Type().Bits())
		if err != nil {
			return err
		}
		rv.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(s, 0, rv.Type().Bits())
		if err != nil {
			return err
		}
		rv.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, rv.Type().Bits())
		if err != nil {
			return err
		}
		rv.SetFloat(v)
	}
	return nil
}
