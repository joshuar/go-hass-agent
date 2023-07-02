// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/mock"
)

type mockSensorRegistry struct {
	mock.Mock
}

func (r *mockSensorRegistry) Open(ctx context.Context, registryPath fyne.URI) error {
	args := r.Called(ctx, registryPath)
	return args.Error(0)
}

func (r *mockSensorRegistry) Get(id string) (*registryItem, error) {
	args := r.Called(id)
	return args.Get(0).(*registryItem), args.Error(1)
}

func (r *mockSensorRegistry) Set(item registryItem) error {
	args := r.Called(item)
	return args.Error(0)
}

func (r *mockSensorRegistry) Close() error {
	args := r.Called()
	return args.Error(0)
}

func newMockSensorTracker(t *testing.T) *SensorTracker {
	fakeRegistry := &mockSensorRegistry{}
	fakeTracker := &SensorTracker{
		sensor:   make(map[string]*sensorState),
		registry: fakeRegistry,
	}
	return fakeTracker
}

func TestSensorTracker_Get(t *testing.T) {
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
		// TODO: Add test cases.
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

func TestSensorTracker_add(t *testing.T) {
	type fields struct {
		registry   Registry
		sensor     map[string]*sensorState
		hassConfig *hass.HassConfig
		mu         sync.RWMutex
	}
	type args struct {
		sensor *sensorState
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
			tracker := &SensorTracker{
				registry:   tt.fields.registry,
				sensor:     tt.fields.sensor,
				hassConfig: tt.fields.hassConfig,
				mu:         tt.fields.mu,
			}
			if err := tracker.add(tt.args.sensor); (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
