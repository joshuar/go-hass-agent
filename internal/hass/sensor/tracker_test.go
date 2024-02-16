// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
)

var mockRegistry = RegistryMock{
	SetDisabledFunc:   func(sensor string, state bool) error { return nil },
	SetRegisteredFunc: func(sensor string, state bool) error { return nil },
	IsDisabledFunc:    func(sensor string) bool { return false },
	IsRegisteredFunc:  func(sensor string) bool { return false },
}

func TestSensorTracker_add(t *testing.T) {
	type fields struct {
		registry Registry
		sensor   map[string]Details
		mu       sync.Mutex
	}
	type args struct {
		s Details
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful add",
			fields: fields{
				sensor: make(map[string]Details),
				mu:     sync.Mutex{},
			},
			args:    args{s: &mockSensor},
			wantErr: false,
		},
		{
			name:    "unsuccessful add (not initialised properly)",
			args:    args{s: &mockSensor},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if err := tr.add(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSensorTracker_Get(t *testing.T) {
	mockMap := make(map[string]Details)
	mockMap["mock_sensor"] = &mockSensor

	type fields struct {
		registry Registry
		sensor   map[string]Details
		mu       sync.Mutex
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Details
		wantErr bool
	}{
		{
			name:    "successful get",
			fields:  fields{sensor: mockMap},
			args:    args{id: "mock_sensor"},
			wantErr: false,
			want:    &mockSensor,
		},
		{
			name:    "unsuccessful get",
			fields:  fields{sensor: mockMap},
			args:    args{id: "doesntExist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			got, err := tr.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorTracker.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorTracker_SensorList(t *testing.T) {
	mockMap := make(map[string]Details)
	mockMap["mock_sensor"] = &mockSensor

	type fields struct {
		registry Registry
		sensor   map[string]Details
		mu       sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "with sensors",
			fields: fields{sensor: mockMap},
			want:   []string{"mock_sensor"},
		},
		{
			name: "without sensors",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if got := tr.SensorList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorTracker.SensorList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSensorTracker(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		want    *Tracker
		wantErr bool
	}{
		{
			name: "default test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTracker()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSensorTracker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewSensorTracker() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestMergeSensorCh(t *testing.T) {
	type args struct {
		ctx      context.Context
		sensorCh []<-chan Details
	}
	tests := []struct {
		name string
		args args
		want chan Details
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeSensorCh(tt.args.ctx, tt.args.sensorCh...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeSensorCh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorTracker_Reset(t *testing.T) {
	mockMap := make(map[string]Details)
	mockMap["mock_sensor"] = &mockSensor

	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "default",
			fields: fields{sensor: mockMap},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			tr.Reset()
			assert.Nil(t, tr.sensor)
		})
	}
}

// func Test_handleUpdates(t *testing.T) {
// 	successful := &updateResponse{"mockSensor": {Success: true}}
// 	unsuccessful := &updateResponse{"mockSensor": {Success: false}}
// 	type args struct {
// 		reg Registry
// 		r   *updateResponse
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name:    "successful update",
// 			args:    args{reg: &mockRegistry, r: successful},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "unsuccessful update",
// 			args:    args{reg: &mockRegistry, r: unsuccessful},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := handleUpdates(tt.args.reg, tt.args.r); (err != nil) != tt.wantErr {
// 				t.Errorf("handleUpdates() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_handleRegistration(t *testing.T) {
// 	type args struct {
// 		reg Registry
// 		r   *registrationResponse
// 		s   string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name:    "successful registration",
// 			args:    args{reg: &mockRegistry, r: &registrationResponse{Success: true}, s: "mockSensor"},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "unsuccessful registration",
// 			args:    args{reg: &mockRegistry, r: &registrationResponse{Success: false}, s: "mockSensor"},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := handleRegistration(tt.args.reg, tt.args.r, tt.args.s); (err != nil) != tt.wantErr {
// 				t.Errorf("handleRegistration() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func Test_handleResponse(t *testing.T) {
	type args struct {
		resp hass.Response
		trk  *Tracker
		upd  Details
		reg  Registry
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleResponse(tt.args.resp, tt.args.trk, tt.args.upd, tt.args.reg)
		})
	}
}

func TestTracker_UpdateSensor(t *testing.T) {
	// setup creates a context using a test http client and server which will
	// return the given response when the ExecuteRequest function is called.
	setup := func(r hass.Response) context.Context {
		ctx := context.TODO()
		// load client
		client := resty.New().
			SetTimeout(1 * time.Second).
			AddRetryCondition(
				func(rr *resty.Response, err error) bool {
					return rr.StatusCode() == http.StatusTooManyRequests
				},
			)
		ctx = hass.ContextSetClient(ctx, client)
		// load server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var resp []byte
			var err error
			switch rType := r.(type) {
			case *updateResponse:
				resp, err = json.Marshal(rType.Body)
			case *registrationResponse:
				resp, err = json.Marshal(rType.Body)
			}
			if err != nil {
				w.Write(json.RawMessage(`{"success":false}`))
			} else {
				w.Write(resp)
			}
		}))
		ctx = hass.ContextSetURL(ctx, server.URL)
		// return loaded context
		return ctx
	}

	// set up a fake sensor tracker
	mockMap := make(map[string]Details)
	// set up a fake registry with sensors
	registry.SetPath(t.TempDir())
	reg, err := registry.Load()
	assert.Nil(t, err)

	// test: disabled sensor
	// - has a old state
	// - has a new state that SHOULD NOT be set
	oldDisabledSensor := mockSensor
	oldDisabledSensor.IDFunc = func() string { return "disabled_sensor" }
	mockMap[oldDisabledSensor.IDFunc()] = &oldDisabledSensor
	err = reg.SetDisabled(oldDisabledSensor.IDFunc(), true)
	assert.Nil(t, err)
	err = reg.SetRegistered(oldDisabledSensor.IDFunc(), true)
	assert.Nil(t, err)
	// a new state update for the disabled sensor
	newDisabledSensor := oldDisabledSensor
	newDisabledSensor.StateFunc = func() any { return "disabledState" }

	// test: new sensor
	// - does not exist in registry
	newSensor := mockSensor
	newSensor.IDFunc = func() string { return "new_sensor" }
	newSensor.StateFunc = func() any { return "newState" }
	newSensorResponse := NewRegistrationResponse()
	newSensorResponse.Body = response{Success: true}

	// test: updated sensor
	// - does not exist in registry
	updatedSensor := mockSensor
	updatedSensor.StateFunc = func() any { return "newState" }
	updatedSensorResponse := NewUpdateResponse()
	updatedSensorResponse.Body[updatedSensor.IDFunc()] = &response{Success: true}
	mockMap[updatedSensor.IDFunc()] = &updatedSensor
	err = reg.SetDisabled(updatedSensor.IDFunc(), false)
	assert.Nil(t, err)
	err = reg.SetRegistered(updatedSensor.IDFunc(), true)
	assert.Nil(t, err)

	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	type args struct {
		ctx context.Context
		reg Registry
		upd Details
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string
	}{
		{
			name:   "disabled sensor",
			fields: fields{sensor: mockMap},
			args:   args{ctx: context.TODO(), reg: reg, upd: &newDisabledSensor},
			want:   "mockState",
		},
		{
			name:   "new sensor",
			fields: fields{sensor: mockMap},
			args:   args{ctx: setup(newSensorResponse), reg: reg, upd: &newSensor},
			want:   "newState",
		},
		{
			name:   "updated sensor",
			fields: fields{sensor: mockMap},
			args:   args{ctx: setup(updatedSensorResponse), reg: reg, upd: &updatedSensor},
			want:   "newState",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if err := tr.UpdateSensor(tt.args.ctx, tt.args.reg, tt.args.upd); (err != nil) != tt.wantErr {
				t.Errorf("Tracker.UpdateSensor() error = %v, wantErr %v", err, tt.wantErr)
			}
			// TODO: also test the tracker has the right state
			// assert.Equal(t, tr.sensor[tt.args.upd.ID()].State(), tt.want)
		})
	}
}
