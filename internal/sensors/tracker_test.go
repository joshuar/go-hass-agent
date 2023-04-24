// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"os"
	"testing"

	"fyne.io/fyne/v2/app"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSensorUpdate struct {
	mock.Mock
}

func (m *mockSensorUpdate) Name() string {
	m.On("Name")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) ID() string {
	m.On("ID")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) Icon() string {
	m.On("Icon")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) SensorType() hass.SensorType {
	m.On("SensorType")
	m.Called()
	return hass.TypeSensor
}

func (m *mockSensorUpdate) DeviceClass() hass.SensorDeviceClass {
	m.On("DeviceClass")
	m.Called()
	return 0
}

func (m *mockSensorUpdate) StateClass() hass.SensorStateClass {
	m.On("StateClass")
	m.Called()
	return 0
}

func (m *mockSensorUpdate) State() interface{} {
	m.On("State")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) Units() string {
	m.On("Units")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) Category() string {
	m.On("Category")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) Attributes() interface{} {
	m.On("Attributes")
	m.Called()
	return nil
}

type MockSensorRegistry struct {
	mock.Mock
}

func newMockSensorRegistry(t *testing.T) *sensorRegistry {
	fakeRegistry := new(sensorRegistry)
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	assert.Nil(t, err)
	fakeRegistry.db = db
	return fakeRegistry
}

func newMockSensorTracker(t *testing.T) *sensorTracker {
	fakeRegistry := newMockSensorRegistry(t)
	fakeTracker := &sensorTracker{
		sensor:        make(map[string]*sensorState),
		sensorWorkers: nil,
		registry:      fakeRegistry,
		hassConfig:    nil,
	}
	return fakeTracker
}

var testApp = app.NewWithID("org.joshuar.go-hass-agent-test")
var uri = testApp.Storage().RootURI()

func TestNewSensorTracker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracker := NewSensorTracker(ctx, uri)
	assert.IsType(t, &sensorTracker{}, tracker)

	os.RemoveAll(uri.Path())
}

func TestAdd(t *testing.T) {
	fakeSensorUpdate := &mockSensorUpdate{}
	tracker := newMockSensorTracker(t)
	err := tracker.add(fakeSensorUpdate)
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	tracker := newMockSensorTracker(t)

	// test non-existent sensor
	got := tracker.get("nonexistent")
	assert.Nil(t, got)

	// test existing sensor
	fakeSensorUpdate := &mockSensorUpdate{}
	err := tracker.add(fakeSensorUpdate)
	assert.Nil(t, err)
	got = tracker.get(fakeSensorUpdate.ID())
	assert.NotNil(t, got)
}

func TestExists(t *testing.T) {

	tracker := newMockSensorTracker(t)

	// test sensor doesn't exist
	input := "NonExistentSensorID"
	assert.False(t, tracker.exists(input))

	// test sensor exists
	fakeSensorUpdate := &mockSensorUpdate{}
	err := tracker.add(fakeSensorUpdate)
	assert.Nil(t, err)
	assert.True(t, tracker.exists(fakeSensorUpdate.ID()))
}
