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
	validConfig := &AgentConfigMock{
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

	invalidConfig := &AgentConfigMock{
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

func Test_generateWebsocketURL(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid host",
			args: args{host: "http://localhost:8123"},
			want: "ws://localhost:8123/api/websocket",
		},
		{
			name: "invalid host",
			args: args{host: "localhost:8123"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateWebsocketURL(tt.args.host); got != tt.want {
				t.Errorf("generateWebsocketURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateAPIURL(t *testing.T) {
	type args struct {
		host         string
		cloudhookURL string
		remoteUIURL  string
		webhookID    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "host and webhook",
			args: args{
				host:      "http://localhost:8123",
				webhookID: "123",
			},
			want: "http://localhost:8123/api/webhook/123",
		},
		{
			name: "cloudhook",
			args: args{
				cloudhookURL: "http://localhost:8123",
			},
			want: "http://localhost:8123",
		},
		{
			name: "remoteuiurl",
			args: args{
				remoteUIURL: "http://localhost:8123",
				webhookID:   "123",
			},
			want: "http://localhost:8123/api/webhook/123",
		},
		{
			name: "host but missing webhook",
			args: args{
				host: "http://localhost:8123",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateAPIURL(tt.args.host, tt.args.cloudhookURL, tt.args.remoteUIURL, tt.args.webhookID); got != tt.want {
				t.Errorf("generateAPIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
