// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:containedctx,exhaustruct,nlreturn,paralleltest,wsl
package hass

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestContextSetClient(t *testing.T) {
	goodClient := resty.New()

	type args struct {
		ctx    context.Context
		client *resty.Client
	}

	tests := []struct {
		args args
		want *resty.Client
		name string
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
		args args
		want *resty.Client
		name string
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

func TestSetupContext(t *testing.T) {
	restAPIURL := "http://localhost:8123/api"
	prefs := preferences.DefaultPreferences()
	prefs.RestAPIURL = restAPIURL
	ctx := preferences.ContextSetPrefs(context.TODO(), prefs)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "with preferences",
			args: args{ctx: ctx},
			want: restAPIURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetupContext(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetupContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			client := ContextGetClient(got)
			assert.NotNil(t, client)
			assert.Equal(t, tt.want, client.BaseURL)
		})
	}
}
