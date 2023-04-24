// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/assert"
)

type fakeSensorData struct {
	name  string
	state interface{}
}

func (f *fakeSensorData) Name() string {
	return f.name
}

func (f *fakeSensorData) ID() string {
	return f.name
}

func (f *fakeSensorData) Icon() string {
	return "mdi:user"
}
func (f *fakeSensorData) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (f *fakeSensorData) DeviceClass() hass.SensorDeviceClass {
	return 0
}

func (f *fakeSensorData) StateClass() hass.SensorStateClass {
	return 0
}

func (f *fakeSensorData) State() interface{} {
	return f.state
}

func (f *fakeSensorData) Units() string {
	return ""
}
func (f *fakeSensorData) Category() string {
	return ""
}

func (f *fakeSensorData) Attributes() interface{} {
	return nil
}

func TestSensorState(t *testing.T) {
	fakeAttributes := &struct {
		MyAttribute string `json:"My Attribute"`
	}{
		MyAttribute: "fakeAttribute",
	}
	fakeMetadata := &sensorMetadata{
		Registered: true,
		Disabled:   false,
	}
	fakeSensorState := &sensorState{
		deviceClass: 0,
		stateClass:  hass.StateMeasurement,
		sensorType:  hass.TypeSensor,
		state:       "fakeValue",
		stateUnits:  "",
		attributes:  fakeAttributes,
		icon:        "mdi:person",
		name:        "Fake Sensor",
		entityID:    "fake_sensor",
		category:    "",
		metadata:    fakeMetadata,
	}

	assert.Equal(t, "", fakeSensorState.DeviceClass())
	assert.Equal(t, hass.StateMeasurement.String(), fakeSensorState.StateClass())
	assert.Equal(t, hass.TypeSensor.String(), fakeSensorState.Type())
	assert.Equal(t, "mdi:person", fakeSensorState.Icon())
	assert.Equal(t, "Fake Sensor", fakeSensorState.Name())
	assert.Equal(t, "fakeValue", fakeSensorState.State())
	assert.Equal(t, fakeAttributes, fakeSensorState.Attributes())
	assert.Equal(t, "fake_sensor", fakeSensorState.UniqueID())
	assert.Equal(t, "", fakeSensorState.UnitOfMeasurement())
	assert.Equal(t, "", fakeSensorState.EntityCategory())
	assert.True(t, fakeSensorState.Registered())
	assert.False(t, fakeSensorState.Disabled())
}

func TestMarshal(t *testing.T) {
	fakeSensorUpdate := &fakeSensorData{
		name:  "FakeSensor",
		state: "FakeValue",
	}
	got := marshalSensorState(fakeSensorUpdate)
	assert.IsType(t, &sensorState{}, got)
}
