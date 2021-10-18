// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"net/url"
)

const (
	queryParamTagName = "queryparam"
)

type DecodeQueryParamOptions struct {
	TagName      string
	TagOnly      bool
	KeyConverter func(string) string
}

// DecodeQueryParam decodes query parameters into a struct.
func DecodeQueryParam(u url.Values, v interface{}, o *DecodeQueryParamOptions) error {
	m := map[string]string{}
	for k, v := range u {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	if o == nil {
		o = &DecodeQueryParamOptions{}
	}
	if o.TagName == "" {
		o.TagName = queryParamTagName
	}
	opts := &DecodeStringMapOptions{
		TagName:      o.TagName,
		TagOnly:      o.TagOnly,
		KeyConverter: o.KeyConverter,
	}
	return DecodeStringMap(m, v, opts)
}
