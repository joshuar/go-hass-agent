// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/api"
)

func Test_haConfig_RequestType(t *testing.T) {
	type fields struct {
		rawConfigProps map[string]interface{}
		haConfigProps  haConfigProps
		mu             sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   api.RequestType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &haConfig{
				rawConfigProps: tt.fields.rawConfigProps,
				haConfigProps:  tt.fields.haConfigProps,
				mu:             tt.fields.mu,
			}
			if got := h.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("haConfig.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_haConfig_RequestData(t *testing.T) {
	type fields struct {
		rawConfigProps map[string]interface{}
		haConfigProps  haConfigProps
		mu             sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   json.RawMessage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &haConfig{
				rawConfigProps: tt.fields.rawConfigProps,
				haConfigProps:  tt.fields.haConfigProps,
				mu:             tt.fields.mu,
			}
			if got := h.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("haConfig.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_haConfig_ResponseHandler(t *testing.T) {
	type fields struct {
		rawConfigProps map[string]interface{}
		haConfigProps  haConfigProps
		mu             sync.Mutex
	}
	type args struct {
		resp   bytes.Buffer
		respCh chan api.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &haConfig{
				rawConfigProps: tt.fields.rawConfigProps,
				haConfigProps:  tt.fields.haConfigProps,
				mu:             tt.fields.mu,
			}
			h.ResponseHandler(tt.args.resp, tt.args.respCh)
		})
	}
}

func Test_getConfig(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *haConfig
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfig(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRegisteredEntities(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRegisteredEntities(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegisteredEntities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRegisteredEntities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEntityDisabled(t *testing.T) {
	type args struct {
		ctx    context.Context
		entity string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsEntityDisabled(tt.args.ctx, tt.args.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsEntityDisabled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsEntityDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVersion(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
