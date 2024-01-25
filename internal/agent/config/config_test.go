// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	_ "embed"
	"errors"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	validConfig := &ConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case PrefAPIURL:
				*v = "http://localhost:8123"
				return nil
			case PrefWebsocketURL:
				*v = "http://localhost:8123"
				return nil
			case PrefToken:
				*v = "123456"
				return nil
			case PrefWebhookID:
				*v = "123456"
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	invalidConfig := &ConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case PrefAPIURL:
				*v = "not a url"
				return nil
			case PrefWebsocketURL:
				*v = "not a url"
				return nil
			case PrefToken:
				*v = ""
				return nil
			case PrefWebhookID:
				*v = ""
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	type args struct {
		c Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid config",
			args:    args{c: validConfig},
			wantErr: false,
		},
		{
			name:    "invalid config",
			args:    args{c: invalidConfig},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateConfig(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
