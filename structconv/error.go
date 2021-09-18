// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"encoding/json"
	"strings"
)

const (
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
