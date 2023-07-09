// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/api"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewHassConfig(t *testing.T) {
	newMockServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &api.UnencryptedRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.Nil(t, err)
			assert.Equal(t, "get_config", req.Type)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
	}
	mockServer := newMockServer(t)
	defer mockServer.Close()

	mockConfig := config.NewMockConfig(t)
	mockConfig.On("Get", "apiURL").Return(mockServer.URL, nil)
	ctx := config.StoreInContext(context.Background(), mockConfig)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *HassConfig
		wantErr bool
	}{
		{
			name:    "successful",
			args:    args{ctx: ctx},
			want:    &HassConfig{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHassConfig(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHassConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewHassConfig() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestHassConfig_GetEntityState(t *testing.T) {
	mockEntities := make(map[string]map[string]interface{})
	mockEntities["aSensor"] = make(map[string]interface{})
	mockEntities["aSensor"]["state"] = "value"
	mockProps := hassConfigProps{
		Entities: mockEntities,
	}

	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
	}
	type args struct {
		entity string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{
			name:   "existing entity",
			fields: fields{hassConfigProps: mockProps},
			args:   args{entity: "aSensor"},
			want:   mockEntities["aSensor"],
		},
		{
			name:   "nonexisting entity",
			fields: fields{hassConfigProps: mockProps},
			args:   args{entity: "notSensor"},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if got := h.GetEntityState(tt.args.entity); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HassConfig.GetEntityState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHassConfig_IsEntityDisabled(t *testing.T) {
	mockEntities := make(map[string]map[string]interface{})
	mockEntities["disabledSensor"] = make(map[string]interface{})
	mockEntities["disabledSensor"]["disabled"] = true
	mockEntities["enabledSensor"] = make(map[string]interface{})
	mockEntities["enabledSensor"]["disabled"] = false
	mockProps := hassConfigProps{
		Entities: mockEntities,
	}

	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
	}
	type args struct {
		entity string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "disabled sensor",
			fields: fields{hassConfigProps: mockProps},
			args:   args{entity: "disabledSensor"},
			want:   true,
		},
		{
			name:   "disabled sensor",
			fields: fields{hassConfigProps: mockProps},
			args:   args{entity: "enabledSensor"},
			want:   false,
		},
		{
			name:   "no props",
			fields: fields{hassConfigProps: hassConfigProps{}},
			args:   args{entity: "disabledSensor"},
			want:   false,
		},
		{
			name: "nil entities",
			fields: fields{hassConfigProps: hassConfigProps{
				Entities: nil,
			}},
			args: args{entity: "disabledSensor"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if got := h.IsEntityDisabled(tt.args.entity); got != tt.want {
				t.Errorf("HassConfig.IsEntityDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHassConfig_RequestType(t *testing.T) {
	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
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
			h := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if got := h.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HassConfig.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHassConfig_RequestData(t *testing.T) {
	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
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
			h := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if got := h.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HassConfig.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHassConfig_ResponseHandler(t *testing.T) {
	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
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
			h := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			h.ResponseHandler(tt.args.resp, tt.args.respCh)
		})
	}
}

func TestHassConfig_Get(t *testing.T) {
	mockProps := make(map[string]interface{})
	mockProps["exists"] = "value"

	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
	}
	type args struct {
		property string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "get existing",
			fields:  fields{rawConfigProps: mockProps},
			args:    args{property: "exists"},
			want:    "value",
			wantErr: false,
		},
		{
			name:    "get nonexisting",
			fields:  fields{rawConfigProps: mockProps},
			args:    args{property: "nonexisting"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			got, err := c.Get(tt.args.property)
			if (err != nil) != tt.wantErr {
				t.Errorf("HassConfig.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HassConfig.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHassConfig_Set(t *testing.T) {
	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
	}
	type args struct {
		property string
		value    interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if err := c.Set(tt.args.property, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("HassConfig.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHassConfig_Validate(t *testing.T) {
	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("HassConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHassConfig_Refresh(t *testing.T) {
	newMockServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &api.UnencryptedRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.Nil(t, err)
			assert.Equal(t, "get_config", req.Type)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
	}
	mockServer := newMockServer(t)
	defer mockServer.Close()

	mockConfig := config.NewMockConfig(t)
	mockConfig.On("Get", "apiURL").Return(mockServer.URL, nil)
	ctx := config.StoreInContext(context.Background(), mockConfig)

	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful",
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name: "unsuccessful",
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if err := h.Refresh(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("HassConfig.Refresh() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHassConfig_Upgrade(t *testing.T) {
	type fields struct {
		rawConfigProps  map[string]interface{}
		hassConfigProps hassConfigProps
		mu              sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HassConfig{
				rawConfigProps:  tt.fields.rawConfigProps,
				hassConfigProps: tt.fields.hassConfigProps,
				mu:              tt.fields.mu,
			}
			if err := h.Upgrade(); (err != nil) != tt.wantErr {
				t.Errorf("HassConfig.Upgrade() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
