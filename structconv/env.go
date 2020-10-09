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
	if err := initStruct(v); err != nil {
		return err
	}
	parser := func(f reflect.StructField) (interface{}, error) {
		return parseDecodeTag(f, envTagName)
	}
	info, err := getStructInfo(v, parser)
	if err != nil {
		return err
	}
	var strMap stringMap = &envData{}
	if err := stringMapToStruct(info, strMap, getEnvKey); err != nil {
		return err
	}
	return nil
}

func getEnvKey(info *fieldInfo, tag decodeTagInfo) string {
	var key string
	if tag.Ok && tag.Key != "" {
		key = tag.Key
	} else {
		key = strcase.ToUpperSnake(info.Meta.Name)
	}
	return key
}
