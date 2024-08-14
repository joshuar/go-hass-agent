// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package preferences

import "testing"

func TestRegistration_Validate(t *testing.T) {
	type fields struct {
		Server string
		Token  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid",
			fields: fields{Server: "http://localhost:8123", Token: "ALongSecretString"},
		},
		{
			name:    "invalid",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Registration{
				Server: tt.fields.Server,
				Token:  tt.fields.Token,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Registration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
