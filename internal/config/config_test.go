// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestStoreInContext(t *testing.T) {
	wantedCtx := context.WithValue(context.Background(),
		configKey,
		NewMockConfig(t))
	type args struct {
		ctx context.Context
		c   Config
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "standard test",
			args: args{ctx: context.Background(), c: NewMockConfig(t)},
			want: wantedCtx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StoreInContext(tt.args.ctx, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StoreInContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchFromContext(t *testing.T) {
	validCtx := context.WithValue(context.Background(),
		configKey,
		NewMockConfig(t))
	invalidCtx := context.Background()
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{
			name:    "fetch valid",
			args:    args{ctx: validCtx},
			want:    NewMockConfig(t),
			wantErr: false,
		},
		{
			name:    "fetch invalid",
			args:    args{ctx: invalidCtx},
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

func TestFetchPropertyFromContext(t *testing.T) {
	config := NewMockConfig(t)
	config.On("Get", "valid").Return("validValue", nil)
	config.On("Get", "invalid").Return("", errors.New("invalid"))

	validCtx := context.WithValue(context.Background(),
		configKey,
		config)
	invalidCtx := context.Background()
	type args struct {
		ctx      context.Context
		property string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "test valid property",
			args:    args{ctx: validCtx, property: "valid"},
			want:    "validValue",
			wantErr: false,
		},
		{
			name:    "test invalid property",
			args:    args{ctx: validCtx, property: "invalid"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "test invalid context",
			args:    args{ctx: invalidCtx, property: "valid"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchPropertyFromContext(tt.args.ctx, tt.args.property)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchPropertyFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchPropertyFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
