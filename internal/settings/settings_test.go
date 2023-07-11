// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package settings

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

func TestSettings_GetValue(t *testing.T) {
	values := make(map[string]string)
	values["aKey"] = "aValue"
	type fields struct {
		mu     sync.RWMutex
		values map[string]string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "existing value",
			fields:  fields{values: values},
			args:    args{key: "aKey"},
			want:    "aValue",
			wantErr: false,
		},
		{
			name:    "missing value",
			fields:  fields{values: values},
			args:    args{key: "notaKey"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Settings{
				mu:     tt.fields.mu,
				values: tt.fields.values,
			}
			got, err := s.GetValue(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Settings.GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Settings.GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettings_SetValue(t *testing.T) {
	type fields struct {
		mu     sync.RWMutex
		values map[string]string
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "successful set",
			fields:  fields{values: make(map[string]string)},
			args:    args{key: "aKey", value: "aValue"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Settings{
				mu:     tt.fields.mu,
				values: tt.fields.values,
			}
			if err := s.SetValue(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Settings.SetValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSettings(t *testing.T) {
	tests := []struct {
		name string
		want *Settings
	}{
		{
			name: "default test",
			want: &Settings{values: make(map[string]string)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSettings(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSettings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStoreInContext(t *testing.T) {
	goodCtx := context.WithValue(context.Background(), contextKey, NewSettings())
	type args struct {
		ctx context.Context
		s   *Settings
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "default test",
			args: args{ctx: context.Background(), s: NewSettings()},
			want: goodCtx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StoreInContext(tt.args.ctx, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StoreInContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchFromContext(t *testing.T) {
	goodCtx := StoreInContext(context.Background(), NewSettings())
	badCtx := context.Background()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *Settings
		wantErr bool
	}{
		{
			name:    "settings exists",
			args:    args{ctx: goodCtx},
			want:    NewSettings(),
			wantErr: false,
		},
		{
			name:    "settings does not exist",
			args:    args{ctx: badCtx},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchFromContext(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
