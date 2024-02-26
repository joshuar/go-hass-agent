// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestContextSetURL(t *testing.T) {
	type args struct {
		ctx context.Context
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "successful test",
			args: args{ctx: context.TODO(), url: "good"},
			want: "good",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContextSetURL(tt.args.ctx, tt.args.url)
			url, ok := got.Value(urlContextKey).(string)
			assert.True(t, ok)
			assert.Equal(t, url, tt.want)
		})
	}
}

func TestContextGetURL(t *testing.T) {
	testCtx := ContextSetURL(context.TODO(), "test")

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "successful test",
			args: args{ctx: testCtx},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContextGetURL(tt.args.ctx); got != tt.want {
				t.Errorf("ContextGetURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextSetClient(t *testing.T) {
	goodClient := resty.New()
	type args struct {
		ctx    context.Context
		client *resty.Client
	}
	tests := []struct {
		name string
		args args
		want *resty.Client
	}{
		{
			name: "successful test",
			args: args{ctx: context.TODO(), client: goodClient},
			want: goodClient,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContextSetClient(tt.args.ctx, tt.args.client)
			client, ok := got.Value(clientContextKey).(*resty.Client)
			assert.True(t, ok)
			if !reflect.DeepEqual(client, tt.want) {
				t.Errorf("ContextSetClient() = %v, want %v", client, tt.want)
			}
		})
	}
}

func TestContextGetClient(t *testing.T) {
	goodClient := resty.New()
	goodCtx := ContextSetClient(context.TODO(), goodClient)
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *resty.Client
	}{
		{
			name: "successful test",
			args: args{ctx: goodCtx},
			want: goodClient,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContextGetClient(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContextGetClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewContext(t *testing.T) {
	mockServer := "http://test.host:9999"
	preferences.SetPath(t.TempDir())
	preferences.Save(
		preferences.SetHost(mockServer),
		preferences.SetToken("testToken"),
		preferences.SetCloudhookURL(""),
		preferences.SetRemoteUIURL(""),
		preferences.SetWebhookID("testID"),
		preferences.SetSecret(""),
		preferences.SetRestAPIURL(mockServer),
		preferences.SetWebsocketURL(mockServer),
		preferences.SetDeviceName("testDevice"),
		preferences.SetDeviceID("testID"),
		preferences.SetVersion("6.4.0"),
		preferences.SetRegistered(true),
	)
	tests := []struct {
		name    string
		wantURL string
	}{
		{
			name:    "successful test",
			wantURL: mockServer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := NewContext()
			gotURL := ContextGetURL(got)
			assert.Equal(t, gotURL, tt.wantURL)
			gotClient := ContextGetClient(got)
			assert.NotNil(t, gotClient)
		})
	}
}
