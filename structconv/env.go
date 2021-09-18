// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"os"
	"reflect"

	"github.com/twihike/go-strcase/strcase"
)

const (
	envTagName = "env"
)

type envData struct{}

func (d *envData) Get(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	return v, ok
}

// DecodeEnv decodes environment variables into a struct.
func DecodeEnv(v interface{}) error {
	s, err := checkStructPtr(v)
	if err != nil {
		return err
	}
	if err := initStruct(v); err != nil {
		return err
	}
	parser := func(f reflect.StructField) (decodeTagInfo, error) {
		return parseDecodeTag(f, envTagName)
	}
	var strMap stringMap = &envData{}
	params := stringMapToStructParams{
		Struct:    s,
		StringMap: strMap,
		TagParser: parser,
		KeyParser: getEnvKey,
	}
	if err := stringMapToStruct(params); err != nil {
		return err
	}
	return nil
}

func getEnvKey(info fieldInfo, tag decodeTagInfo) string {
	var key string
	if tag.OK && tag.Key != "" {
		key = tag.Key
	} else {
		key = strcase.ToUpperSnake(info.Meta.Name)
	}
	return key
}
