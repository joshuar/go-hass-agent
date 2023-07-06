// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/joshuar/go-hass-agent/internal/hass"
)

func TestNewSensorTracker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type args struct {
		ctx          context.Context
		registryPath string
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name: "successful create",
			args: args{
				ctx:          ctx,
				registryPath: "",
			},
			wantNil: false,
		},
		{
			name: "unsuccessful create",
			args: args{
				ctx:          ctx,
				registryPath: "/nonexistent",
			},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSensorTracker(tt.args.ctx, tt.args.registryPath)
			spew.Dump(r)
			if (r == nil) != tt.wantNil {
				t.Error("NewSensorTracker() = nil, want *SensorTracker")
			}
		})
	}
}

// func TestSensorTracker_add(t *testing.T) {
// 	registry := NewMockRegistry(t)
// 	registry.On("Set", mock.AnythingOfType("RegistryItem")).Return(nil)
// 	state := mocks.NewSensorUpdate(t)
// 	state.On("ID").Return("sensorID")
// 	type fields struct {
// 		registry   Registry
// 		sensor     map[string]*sensorState
// 		hassConfig *hass.HassConfig
// 		mu         sync.RWMutex
// 	}
// 	type args struct {
// 		sensor *sensorState
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "new sensor",
// 			fields: fields{
// 				registry: registry,
// 				sensor:   make(map[string]*sensorState),
// 			},
// 			args: args{
// 				sensor: &sensorState{
// 					data: state,
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "incorrect initialisation",
// 			fields: fields{
// 				registry: registry,
// 			},
// 			args: args{
// 				sensor: &sensorState{
// 					data: state,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tracker := &SensorTracker{
// 				registry: tt.fields.registry,
// 				sensor:   tt.fields.sensor,
// 				mu:       tt.fields.mu,
// 			}
// 			if err := tracker.add(tt.args.sensor); (err != nil) != tt.wantErr {
// 				t.Errorf("SensorTracker.add() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestSensorTracker_Get(t *testing.T) {
	sensors := make(map[string]*sensorState)
	sensors["existingID"] = &sensorState{}
	type fields struct {
		registry Registry
		sensor   map[string]*sensorState
		mu       sync.RWMutex
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    sensorState
		wantErr bool
	}{
		{
			name: "sensor exists",
			fields: fields{
				sensor: sensors,
			},
			args: args{
				id: "existingID",
			},
			wantErr: false,
		},
		{
			name: "sensor does not exist",
			fields: fields{
				sensor: sensors,
			},
			args: args{
				id: "nonexistentID",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			got, err := tracker.Get(tt.args.id)
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

func TestSensorTracker_StartWorkers(t *testing.T) {
	type fields struct {
		registry Registry
		sensor   map[string]*sensorState
		mu       sync.RWMutex
	}
	type args struct {
		ctx      context.Context
		updateCh chan interface{}
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
			tracker := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			tracker.StartWorkers(tt.args.ctx, tt.args.updateCh)
		})
	}
}

func TestSensorTracker_Update(t *testing.T) {
	type fields struct {
		registry Registry
		sensor   map[string]*sensorState
		mu       sync.RWMutex
	}
	type args struct {
		ctx context.Context
		s   hass.Sensor
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
			tracker := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			tracker.Update(tt.args.ctx, tt.args.s)
		})
	}
}
