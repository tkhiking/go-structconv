// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"os"
	"testing"
)

func TestDecodeEnv(t *testing.T) {
	type envTest struct {
		String   string
		Bool     bool
		Int      int
		Float64  float64
		Default  string
		Rename   string `env:"ENV_NAME"`
		Required string `env:",required"`
		Omitted  string `env:"-"`
	}

	tests := []struct {
		name string
		in   map[string]string
		want envTest
	}{
		{
			"normal",
			map[string]string{
				"STRING":   "str",
				"BOOL":     "true",
				"INT":      "1",
				"FLOAT64":  "0.3",
				"ENV_NAME": "e",
				"REQUIRED": "r",
				"OMITTED":  "-",
			},
			envTest{"str", true, 1, 0.3, "d", "e", "r", ""},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			os.Clearenv()
			for k, v := range tt.in {
				os.Setenv(k, v)
			}
			var got envTest
			got.Default = "d"
			err := DecodeEnv(&got, nil)
			if err != nil {
				t.Error(err)
			}
			if got != tt.want {
				t.Errorf("\nwant = %+v\ngot  = %+v", tt.want, got)
			}
		})
	}
}
