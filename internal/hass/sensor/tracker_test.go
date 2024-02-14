// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
)

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
			tr := &SensorTracker{
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
			tr := &SensorTracker{
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
			tr := &SensorTracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if got := tr.SensorList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorTracker.SensorList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorTracker_UpdateSensor(t *testing.T) {
	// set up a fake sensor tracker
	mockMap := make(map[string]Details)
	// set up a fake registry with sensors
	registry.SetPath(t.TempDir())
	reg, err := registry.Load()
	assert.Nil(t, err)

	// updated sensor with new state
	updatedSensor := mockSensor
	updatedSensor.StateFunc = func() any { return "newState" }
	mockMap["mock_sensor"] = &mockSensor
	err = reg.SetRegistered("mock_sensor", true)
	assert.Nil(t, err)

	// new sensor
	newSensor := mockSensor
	newSensor.IDFunc = func() string { return "new_sensor" }
	newSensor.StateFunc = func() any { return "newState" }

	// disabled sensor
	disabledSensor := mockSensor
	disabledSensor.IDFunc = func() string { return "disabled_sensor" }
	mockMap["disabled_sensor"] = &disabledSensor
	err = reg.SetDisabled("disabled_sensor", true)
	assert.Nil(t, err)
	err = reg.SetRegistered("disabled_sensor", true)
	assert.Nil(t, err)
	// a new state update for the disabled sensor
	newDisabledSensor := disabledSensor
	newDisabledSensor.StateFunc = func() any { return "disabledState" }

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
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "disabled sensor",
			fields: fields{sensor: mockMap},
			args:   args{ctx: context.TODO(), reg: reg, upd: &newDisabledSensor},
			want:   "mockState",
		},
		// {
		// 	name:   "new sensor",
		// 	fields: fields{sensor: mockMap},
		// 	args:   args{ctx: context.TODO(), reg: reg, upd: &newSensor},
		// 	want:   "newState",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &SensorTracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			tr.UpdateSensor(tt.args.ctx, tt.args.reg, tt.args.upd)
			assert.Equal(t, tr.sensor[tt.args.upd.ID()].State(), tt.want)
		})
	}
}

func TestNewSensorTracker(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		want    *SensorTracker
		wantErr bool
	}{
		{
			name: "default test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSensorTracker()
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

func Test_prettyPrintState(t *testing.T) {
	type args struct {
		s Details
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prettyPrintState(tt.args.s); got != tt.want {
				t.Errorf("prettyPrintState() = %v, want %v", got, tt.want)
			}
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

func TestSensorTracker_handleUpdates(t *testing.T) {
	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	type args struct {
		reg Registry
		r   *UpdateResponse
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
			tr := &SensorTracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if err := tr.handleUpdates(tt.args.reg, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.handleUpdates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSensorTracker_handleRegistration(t *testing.T) {
	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	type args struct {
		reg Registry
		r   *RegistrationResponse
		s   string
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
			tr := &SensorTracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if err := tr.handleRegistration(tt.args.reg, tt.args.r, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.handleRegistration() error = %v, wantErr %v", err, tt.wantErr)
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
			tr := &SensorTracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			tr.Reset()
			assert.Nil(t, tr.sensor)
		})
	}
}
