// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"net/url"
)

const (
	formTagName = "form"
)

type DecodeFormOptions struct {
	TagName      string
	TagOnly      bool
	KeyConverter func(string) string
}

// DecodeForm decodes the form data into a struct.
func DecodeForm(u url.Values, v interface{}, o *DecodeFormOptions) error {
	m := map[string]string{}
	for k, v := range u {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	if o == nil {
		o = &DecodeFormOptions{}
	}
	if o.TagName == "" {
		o.TagName = formTagName
	}
	opts := &DecodeStringMapOptions{
		TagName:      o.TagName,
		TagOnly:      o.TagOnly,
		KeyConverter: o.KeyConverter,
	}
	return DecodeStringMap(m, v, opts)
}
