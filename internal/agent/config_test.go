// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"errors"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
)

func TestValidateConfig(t *testing.T) {

	validConfig := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefAPIURL:
				*v = "http://localhost:8123"
				return nil
			case config.PrefWebsocketURL:
				*v = "http://localhost:8123"
				return nil
			case config.PrefToken:
				*v = "123456"
				return nil
			case config.PrefWebhookID:
				*v = "123456"
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	invalidConfig := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefAPIURL:
				*v = "not a url"
				return nil
			case config.PrefWebsocketURL:
				*v = "not a url"
				return nil
			case config.PrefToken:
				*v = ""
				return nil
			case config.PrefWebhookID:
				*v = ""
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	type args struct {
		c AgentConfig
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

func TestUpgradeConfig(t *testing.T) {
	validConfig := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefVersion:
				*v = "v999.0.0"
				return nil
			default:
				return errors.New("not found")
			}
		},
		SetFunc: func(s string, ifaceVal interface{}) error {
			return nil
		},
	}

	type args struct {
		c AgentConfig
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
		// TODO: test each version upgrade?
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpgradeConfig(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("UpgradeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateWebsocketURL(t *testing.T) {
	validConfig := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefHost:
				*v = "http://localhost:8123"
				return nil
			default:
				return errors.New("not found")
			}
		},
		SetFunc: func(s string, ifaceVal interface{}) error {
			return nil
		},
	}

	type args struct {
		c AgentConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "default test",
			args:    args{c: validConfig},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateWebsocketURL(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("generateWebsocketURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateAPIURL(t *testing.T) {
	validConfig := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefHost:
				*v = "http://localhost:8123"
				return nil
			case config.PrefCloudhookURL:
				*v = "http://localhost:8123"
				return nil
			case config.PrefRemoteUIURL:
				*v = "http://localhost:8123"
				return nil
			case config.PrefWebhookID:
				*v = "123456"
				return nil
			default:
				return errors.New("not found")
			}
		},
		SetFunc: func(s string, ifaceVal interface{}) error {
			return nil
		},
	}

	type args struct {
		c AgentConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "default test",
			args:    args{c: validConfig},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateAPIURL(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("generateAPIURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
