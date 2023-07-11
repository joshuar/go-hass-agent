// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"testing"

	"fyne.io/fyne/v2/data/binding"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationDetails_Validate(t *testing.T) {
	type fields struct {
		Server string
		Token  string
		Device hass.DeviceInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "hostname",
			fields: fields{
				Server: "localhost",
				Token:  "abcde.abcde_abcde",
			},
			want: false,
		},
		{
			name: "hostname and port",
			fields: fields{
				Server: "localhost:8123",
				Token:  "abcde.abcde_abcde",
			},
			want: false,
		},
		{
			name: "url",
			fields: fields{
				Server: "http://localhost",
				Token:  "abcde.abcde_abcde",
			},
			want: true,
		},
		{
			name: "url with port",
			fields: fields{
				Server: "http://localhost:8123",
				Token:  "abcde.abcde_abcde",
			},
			want: true,
		},
		{
			name: "url with trailing slash",
			fields: fields{
				Server: "http://localhost/",
				Token:  "abcde.abcde_abcde",
			},
			want: true,
		},
		{
			name: "invalid url",
			fields: fields{
				Server: "asdegasg://localhost//",
				Token:  "abcde.abcde_abcde",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			server := binding.NewString()
			err = server.Set(tt.fields.Server)
			assert.Nil(t, err)
			token := binding.NewString()
			err = token.Set(tt.fields.Server)
			assert.Nil(t, err)

			r := &RegistrationDetails{
				serverBinding: server,
				tokenBinding:  token,
			}
			if got := r.Validate(); got != tt.want {
				t.Errorf("RegistrationDetails.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
