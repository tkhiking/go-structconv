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

type DecodeStringMapOptions struct {
	TagName      string
	TagOnly      bool
	KeyConverter func(string) string
}

type stringMapToStructParams struct {
	Struct    reflect.Value
	StringMap map[string]string
	Options   DecodeStringMapOptions
}

func nilKeyConverter(s string) string { return s }

// DecodeStringMap decodes a string map into a struct.
func DecodeStringMap(m map[string]string, v interface{}, o *DecodeStringMapOptions) error {
	opts := initDecodeStringMapOptions(o)
	s, err := checkStructPtr(v)
	if err != nil {
		return err
	}
	if err := initStruct(v); err != nil {
		return err
	}
	params := stringMapToStructParams{
		Struct:    s,
		StringMap: m,
		Options:   opts,
	}
	if err := stringMapToStruct(params); err != nil {
		return err
	}
	return nil
}

func initDecodeStringMapOptions(o *DecodeStringMapOptions) DecodeStringMapOptions {
	var result DecodeStringMapOptions
	if o != nil {
		result = *o
	}
	if result.TagName == "" {
		result.TagName = stringMapTagName
	}
	if result.KeyConverter == nil {
		result.KeyConverter = nilKeyConverter
	}
	return result
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
				Options:   params.Options,
			}
			childErrs := doStringMapToStruct(p)
			if len(childErrs) > 0 {
				errs = append(errs, childErrs...)
			}
			return
		}
		tag, err := parseDecodeTag(inf.Meta, params.Options.TagName)
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
		if params.Options.TagOnly && !tag.OK {
			return
		}

		key := getStringMapKey(inf, tag, params.Options.KeyConverter)
		if val, ok := params.StringMap[key]; ok {
			if err := convertStringToField(inf.Meta, inf.Value, val); err != nil {
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

func getStringMapKey(info fieldInfo, tag decodeTagInfo, fn func(string) string) string {
	var key string
	if tag.OK && tag.Key != "" {
		key = tag.Key
	} else {
		key = fn(info.Meta.Name)
	}
	return key
}

func convertStringToField(rf reflect.StructField, rv reflect.Value, in string) error {
	crt := rf.Type
	crv := rv
	var rootPtr *reflect.Value
	for isRoot := true; crt.Kind() == reflect.Ptr; {
		ptr := reflect.New(crt.Elem())
		if isRoot {
			rootPtr = &ptr
			isRoot = false
		} else {
			crv.Set(ptr)
		}
		crt = crt.Elem()
		crv = ptr.Elem()
	}
	if err := doConvertStringToField(crv, in); err != nil {
		return err
	}
	if rootPtr != nil {
		rv.Set(*rootPtr)
	}
	return nil
}

func doConvertStringToField(rv reflect.Value, s string) error {
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
