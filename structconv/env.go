// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"os"
	"strings"

	"github.com/twihike/go-strcase/strcase"
)

const (
	envTagName = "env"
)

type DecodeEnvOptions struct {
	TagName      string
	TagOnly      bool
	KeyConverter func(string) string
}

// DecodeEnv decodes environment variables into a struct.
func DecodeEnv(v interface{}, o *DecodeEnvOptions) error {
	m := map[string]string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			m[pair[0]] = pair[1]
		}
	}
	if o == nil {
		o = &DecodeEnvOptions{}
	}
	if o.TagName == "" {
		o.TagName = envTagName
	}
	if o.KeyConverter == nil {
		o.KeyConverter = strcase.ToUpperSnake
	}
	opts := &DecodeStringMapOptions{
		TagName:      o.TagName,
		TagOnly:      o.TagOnly,
		KeyConverter: o.KeyConverter,
	}
	return DecodeStringMap(m, v, opts)
}
