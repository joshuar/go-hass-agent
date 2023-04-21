// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FakeSensor struct {
	name  string
	state interface{}
}

func (f *FakeSensor) Name() string {
	return f.name
}

func (f *FakeSensor) ID() string {
	return f.name
}

func (f *FakeSensor) Icon() string {
	return "mdi:user"
}
func (f *FakeSensor) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (f *FakeSensor) DeviceClass() hass.SensorDeviceClass {
	return 0
}

func (f *FakeSensor) StateClass() hass.SensorStateClass {
	return 0
}

func (f *FakeSensor) State() interface{} {
	return f.state
}

func (f *FakeSensor) Units() string {
	return ""
}
func (f *FakeSensor) Category() string {
	return ""
}

func (f *FakeSensor) Attributes() interface{} {
	return nil
}

type MockSensorRegistry struct {
	mock.Mock
	uri string
	db  *badger.DB
}

func (m *MockSensorRegistry) Add(id string) *MockRegistryEntry {
	return &MockRegistryEntry{}
}

type MockRegistryEntry struct {
	mock.Mock
}

func TestExistsSuccess(t *testing.T) {
	input := "NonExistentSensorID"
	tracker := &sensorTracker{
		sensor:        make(map[string]*sensorState),
		sensorWorkers: nil,
		registry:      nil,
		hassConfig:    nil,
	}
	if tracker.exists(input) {
		t.Error("Fake sensor should not exist!")
	}
}

func TestAdd(t *testing.T) {
	fakeSensor := &FakeSensor{
		name:  "FakeSensor",
		state: "FakeValue",
	}

	fakeRegistry := new(sensorRegistry)
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	assert.Nil(t, err)
	fakeRegistry.db = db

	tracker := &sensorTracker{
		sensor:        make(map[string]*sensorState),
		sensorWorkers: nil,
		registry:      fakeRegistry,
		hassConfig:    nil,
	}
	tracker.Add(fakeSensor)
	if !tracker.exists("FakeSensor") {
		t.Error("Fake sensor was not added!")
	}
}
