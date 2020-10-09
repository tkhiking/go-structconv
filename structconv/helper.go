// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

const (
	requiredTagValue  = "required"
	msgDetailRequired = "%v is required"
)

// DecodeError is the decoding error information.
type DecodeError struct {
	Message string
	Detail  []*DecodeFieldError
}

func (e *DecodeError) Error() string {
	if e.Message == "" {
		e.Message = "decoding failed"
	}

	var sb strings.Builder
	sb.WriteString("structconv:\n")
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return e.Message
	}
	sb.Write(b)
	return sb.String()
}

// DecodeFieldError is the single field information of DecodeError.
type DecodeFieldError struct {
	Name     string
	Value    string
	Messages []string
}

func (e *DecodeFieldError) Error() string {
	b, err := json.Marshal(e)
	if err != nil {
		return e.Name
	}
	return string(b)
}

type structInfo struct {
	Value  reflect.Value
	Fields map[string]*fieldInfo
	Parent *structInfo
}

type fieldInfo struct {
	Meta        reflect.StructField
	Value       reflect.Value
	Tag         interface{}
	Child       *structInfo
	Collections []reflect.Type
}

type decodeTagInfo struct {
	Ok       bool
	Key      string
	Required bool
	Omitted  bool
}

type tagParser func(reflect.StructField) (interface{}, error)

// initStruct initializes struct pointer.
func initStruct(structPtr interface{}) error {
	pv := reflect.ValueOf(structPtr)
	if pv.Kind() != reflect.Ptr {
		err := errors.New("structconv: structPtr must be a struct pointer")
		return err
	}

	sv := pv.Elem()
	if sv.Kind() != reflect.Struct {
		err := errors.New("structconv: structPtr must be a struct pointer")
		return err
	}

	for i := 0; i < sv.NumField(); i++ {
		fv := sv.Field(i)

		if !fv.CanSet() {
			continue
		}

		// Follow the pointer.
		for fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				if fv.Type().Elem().Kind() != reflect.Struct {
					break
				}
				// Initialize struct pointer.
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			fv = fv.Elem()
		}

		if fv.Kind() == reflect.Struct {
			_ = initStruct(fv.Addr().Interface())
		}
	}
	return nil
}

// getStructInfo gets the struct reflect information.
func getStructInfo(structPtr interface{}, parseTag tagParser) (*structInfo, error) {
	pv := reflect.ValueOf(structPtr)
	if pv.Kind() != reflect.Ptr {
		err := errors.New("structconv: structPtr must be a struct pointer")
		return nil, err
	}

	sv := pv.Elem()
	st := sv.Type()
	if sv.Kind() != reflect.Struct {
		err := errors.New("structconv: structPtr must be a struct pointer")
		return nil, err
	}

	sInfo := &structInfo{
		Value:  sv,
		Parent: nil,
	}
	fInfo := make(map[string]*fieldInfo, sv.NumField())
	for i := 0; i < sv.NumField(); i++ {
		fv := sv.Field(i)
		fm := st.Field(i)

		if !fv.CanSet() {
			continue
		}

		// Follow the pointer.
		tmpfv := fv
		for tmpfv.Kind() == reflect.Ptr {
			if tmpfv.IsNil() {
				break
			}
			tmpfv = tmpfv.Elem()
		}
		// Get child struct.
		var child *structInfo
		if tmpfv.Kind() == reflect.Struct {
			child, _ = getStructInfo(tmpfv.Addr().Interface(), parseTag)
			child.Parent = sInfo
		}
		collections := followStructCollections(fv)

		tag, err := parseTag(fm)
		if err != nil {
			return sInfo, err
		}

		fi := &fieldInfo{
			Meta:        fm,
			Value:       fv,
			Tag:         tag,
			Child:       child,
			Collections: collections,
		}
		fInfo[fm.Name] = fi
	}
	sInfo.Fields = fInfo

	return sInfo, nil
}

func followStructCollections(rv reflect.Value) []reflect.Type {
	var collections []reflect.Type
	rt := rv.Type()
	for {
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			collections = append(collections, rt)
		case reflect.Ptr:
			if rt.Elem().Kind() != reflect.Struct {
				return []reflect.Type{}
			}
		case reflect.Struct:
			return collections
		default:
			return []reflect.Type{}
		}
		rt = rt.Elem()
	}
}

func parseDecodeTag(f reflect.StructField, tagName string) (decodeTagInfo, error) {
	var result decodeTagInfo
	tagStr, ok := f.Tag.Lookup(tagName)
	if !ok {
		return result, nil
	}
	result.Ok = true

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
		}
	}
	return result, nil
}
