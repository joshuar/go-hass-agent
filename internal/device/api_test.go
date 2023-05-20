// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockAPI struct {
	mock.Mock
}

func (m *mockAPI) SensorWorkers() []func(context.Context, chan interface{}) {
	args := m.Called()
	return args.Get(0).([]func(context.Context, chan interface{}))
}

func (m *mockAPI) EndPoint(endpoint string) interface{} {
	m.On("EndPoint", "validEndpoint").Return("valid").Once()
	m.On("EndPoint", "invalidEndpoint").Return("").Once()
	args := m.Called(endpoint)
	return args.String(0)
}

func TestStoreAPIInContext(t *testing.T) {
	wantedCtx := context.WithValue(context.Background(),
		configKey,
		&mockAPI{})
	type args struct {
		ctx context.Context
		a   API
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "standard test",
			args: args{ctx: context.Background(), a: &mockAPI{}},
			want: wantedCtx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StoreAPIInContext(tt.args.ctx, tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StoreAPIInContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchAPIFromContext(t *testing.T) {
	validCtx := context.WithValue(context.Background(),
		configKey,
		&mockAPI{})
	invalidCtx := context.Background()
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    API
		wantErr bool
	}{
		{
			name:    "fetch valid",
			args:    args{ctx: validCtx},
			want:    &mockAPI{},
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
			got, err := FetchAPIFromContext(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAPIFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchAPIFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAPIEndpoint(t *testing.T) {
	api := new(mockAPI)
	type args struct {
		api      API
		endpoint string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid get",
			args: args{api: api, endpoint: "validEndpoint"},
			want: "valid",
		},
		{
			name: "invalid get",
			args: args{api: api, endpoint: "invalidEndpoint"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAPIEndpoint[string](tt.args.api, tt.args.endpoint); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAPIEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
