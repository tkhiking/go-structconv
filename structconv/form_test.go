// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"net/url"
	"testing"
)

func TestDecodeForm(t *testing.T) {
	type formTest struct {
		String   string
		Bool     bool
		Int      int
		Float64  float64
		Default  string
		Rename   string `form:"q"`
		Required string `form:",required"`
		Omitted  string `form:"-"`
	}

	tests := []struct {
		name string
		in   string
		want formTest
	}{
		{
			"normal",
			"String=str&Bool=true&Int=1&Float64=0.3&q=a&Required=r&Omitted=-",
			formTest{"str", true, 1, 0.3, "d", "a", "r", ""},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u, _ := url.ParseQuery(tt.in)
			var got formTest
			got.Default = "d"
			err := DecodeForm(u, &got, nil)
			if err != nil {
				t.Error(err)
			}
			if got != tt.want {
				t.Errorf("\nwant = %+v\ngot  = %+v", tt.want, got)
			}
		})
	}
}
