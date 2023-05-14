// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"context"
	"reflect"
	"testing"

	"fyne.io/fyne/v2"
)

func TestStoreConfigInContext(t *testing.T) {
	baseCtx := context.Background()
	mockConfig := &AppConfig{}
	wantCtx := context.WithValue(baseCtx, configKey, mockConfig)
	type args struct {
		ctx context.Context
		c   *AppConfig
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "standard test",
			args: args{ctx: baseCtx, c: mockConfig},
			want: wantCtx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StoreConfigInContext(tt.args.ctx, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchConfigFromContext(t *testing.T) {
	goodCtx := context.WithValue(context.Background(), configKey, &AppConfig{})
	badCtx := context.Background()
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *AppConfig
		wantErr bool
	}{
		{
			name:    "valid context",
			args:    args{ctx: goodCtx},
			want:    &AppConfig{},
			wantErr: false,
		},
		{
			name:    "invalid context",
			args:    args{ctx: badCtx},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchConfigFromContext(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchConfigFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchConfigFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_Validate(t *testing.T) {
	type fields struct {
		APIURL       string
		WebSocketURL string
		Secret       string
		Token        string
		WebhookID    string
		NotifyCh     chan fyne.Notification
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid config",
			fields: fields{
				APIURL:       "string",
				WebSocketURL: "string",
				Token:        "string",
				WebhookID:    "string",
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			fields: fields{
				APIURL:       "string",
				WebSocketURL: "string",
				WebhookID:    "string",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			fields:  fields{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AppConfig{
				APIURL:       tt.fields.APIURL,
				WebSocketURL: tt.fields.WebSocketURL,
				Secret:       tt.fields.Secret,
				Token:        tt.fields.Token,
				WebhookID:    tt.fields.WebhookID,
				NotifyCh:     tt.fields.NotifyCh,
			}
			if err := config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("AppConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
