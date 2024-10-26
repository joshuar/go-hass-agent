// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package api

import "testing"

func Test_ResponseError_Error(t *testing.T) {
	type fields struct {
		Code    any
		Message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "valid error",
			fields: fields{Code: 404, Message: "Not Found"},
			want:   "code 404: Not Found",
		},
		{
			name:   "no code",
			fields: fields{Message: "Not Found"},
			want:   "Not Found",
		},
		{
			name:   "no message",
			fields: fields{Code: "404"},
			want:   "code 404",
		},
		{
			name: "empty",
			want: "unknown error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ResponseError{
				Code:    tt.fields.Code,
				Message: tt.fields.Message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("apiError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
