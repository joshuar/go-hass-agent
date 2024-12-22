// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package sensor

import (
	"reflect"
	"testing"
)

func requestsEqual(t *testing.T, got, want *Request) bool {
	t.Helper()
	switch {
	case !reflect.DeepEqual(got.RequestType, want.RequestType):
		t.Error("request type does not match")
		return false
	case !reflect.DeepEqual(got.Data, want.Data):
		t.Error("request data does not match")
		return false
	}
	return true
}

func TestNewRequest(t *testing.T) {
	mockEntity := NewSensor(
		WithName("Mock Entity"),
		WithID("mock_entity"),
		WithState(
			WithValue("value"),
		),
	)

	mockLocation := Location{
		Gps: []float64{0.1, 0.2},
	}

	type args struct {
		options []Option[Request]
	}
	tests := []struct {
		name string
		args args
		want *Request
	}{
		{
			name: "registration",
			args: args{options: []Option[Request]{AsSensorRegistration(mockEntity)}},
			want: &Request{
				RequestType: requestTypeRegisterSensor,
				Data: &registrationRequestBody{
					State:      "value",
					ID:         "mock_entity",
					Name:       "Mock Entity",
					EntityType: "sensor",
				},
			},
		},
		{
			name: "update",
			args: args{options: []Option[Request]{AsSensorUpdate(mockEntity)}},
			want: &Request{
				RequestType: requestTypeUpdateSensor,
				Data: &stateRequestBody{
					State:      "value",
					ID:         "mock_entity",
					EntityType: "sensor",
				},
			},
		},
		{
			name: "location",
			args: args{options: []Option[Request]{AsLocationUpdate(mockLocation)}},
			want: &Request{
				RequestType: requestTypeLocation,
				Data: &Location{
					Gps: []float64{0.1, 0.2},
				},
			},
		},
		{
			name: "retryable",
			args: args{options: []Option[Request]{AsSensorUpdate(mockEntity), AsRetryable(true)}},
			want: &Request{
				RequestType: requestTypeUpdateSensor,
				Data: &stateRequestBody{
					State:      "value",
					ID:         "mock_entity",
					EntityType: "sensor",
				},
				retryable: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRequest(tt.args.options...); !requestsEqual(t, got, tt.want) {
				t.Errorf("NewRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
