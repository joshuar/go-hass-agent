// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import "encoding/json"

// SensorUpdate represents an update for a sensor. It reflects the current state
// of the sensor at the point in time it is used. It provides a bridge between
// platform/device and HA implementations of what a sensor is.
type SensorUpdate interface {
	Name() string
	ID() string
	Icon() string
	SensorType() SensorType
	DeviceClass() SensorDeviceClass
	StateClass() SensorStateClass
	State() interface{}
	Units() string
	Category() string
	Attributes() interface{}
}

// Sensor represents a sensor in HA. As an interface, it leaves it up to the
// underlying struct to provide the appropriate data for this representation.
type Sensor interface {
	SensorUpdate
	Disabled() bool
	Registered() bool
	MarshalJSON() ([]byte, error)
	UnMarshalJSON([]byte) error
}

// sensorRegistrationInfo is the JSON structure required to register a sensor
// with HA.
type sensorRegistrationInfo struct {
	StateAttributes   interface{} `json:"attributes,omitempty"`
	DeviceClass       string      `json:"device_class,omitempty"`
	Icon              string      `json:"icon,omitempty"`
	Name              string      `json:"name"`
	State             interface{} `json:"state"`
	Type              string      `json:"type"`
	UniqueID          string      `json:"unique_id"`
	UnitOfMeasurement string      `json:"unit_of_measurement,omitempty"`
	StateClass        string      `json:"state_class,omitempty"`
	EntityCategory    string      `json:"entity_category,omitempty"`
	Disabled          bool        `json:"disabled,omitempty"`
}

// sensorUpdateInfo is the JSON structure required to update HA with the current
// sensor state.
type sensorUpdateInfo struct {
	StateAttributes interface{} `json:"attributes,omitempty"`
	Icon            string      `json:"icon,omitempty"`
	State           interface{} `json:"state"`
	Type            string      `json:"type"`
	UniqueID        string      `json:"unique_id"`
}

func MarshalSensorUpdate(s Sensor) *sensorUpdateInfo {
	var sensorType string
	if s.SensorType() == 0 {
		sensorType = "sensor"
	} else {
		sensorType = s.SensorType().String()
	}
	return &sensorUpdateInfo{
		StateAttributes: s.Attributes(),
		Icon:            s.Icon(),
		State:           s.State(),
		Type:            sensorType,
		UniqueID:        s.ID(),
	}
}

func MarshalSensorRegistration(s Sensor) *sensorRegistrationInfo {
	var deviceClass, stateClass, sensorType string
	if s.DeviceClass() == 0 {
		deviceClass = ""
	} else {
		deviceClass = s.DeviceClass().String()
	}
	if s.StateClass() == 0 {
		stateClass = ""
	} else {
		stateClass = s.StateClass().String()
	}
	if s.SensorType() == 0 {
		sensorType = "sensor"
	} else {
		sensorType = s.SensorType().String()
	}
	return &sensorRegistrationInfo{
		StateAttributes:   s.Attributes(),
		DeviceClass:       deviceClass,
		Icon:              s.Icon(),
		Name:              s.Name(),
		State:             s.State(),
		Type:              sensorType,
		UniqueID:          s.ID(),
		UnitOfMeasurement: s.Units(),
		StateClass:        stateClass,
		EntityCategory:    s.Category(),
		Disabled:          s.Disabled(),
	}
}

// MarshalSensorData takes the sensor data and returns the appropriate JSON
// structure to either register or update the sensor in HA.
func MarshalSensorData(s Sensor) *json.RawMessage {
	var deviceClass, stateClass, sensorType string
	if s.DeviceClass() == 0 {
		deviceClass = ""
	} else {
		deviceClass = s.DeviceClass().String()
	}
	if s.StateClass() == 0 {
		stateClass = ""
	} else {
		stateClass = s.StateClass().String()
	}
	if s.SensorType() == 0 {
		sensorType = "sensor"
	} else {
		sensorType = s.SensorType().String()
	}
	var marshaled []byte
	var raw json.RawMessage
	if s.Registered() {
		d := []sensorUpdateInfo{{
			StateAttributes: s.Attributes(),
			Icon:            s.Icon(),
			State:           s.State(),
			Type:            sensorType,
			UniqueID:        s.ID(),
		},
		}
		marshaled, _ = json.Marshal(d)
		raw = json.RawMessage(marshaled)
		return &raw
	} else {
		d := sensorRegistrationInfo{
			StateAttributes:   s.Attributes(),
			DeviceClass:       deviceClass,
			Icon:              s.Icon(),
			Name:              s.Name(),
			State:             s.State(),
			Type:              sensorType,
			UniqueID:          s.ID(),
			UnitOfMeasurement: s.Units(),
			StateClass:        stateClass,
			EntityCategory:    s.Category(),
			Disabled:          s.Disabled(),
		}
		marshaled, _ = json.Marshal(d)
		raw = json.RawMessage(marshaled)
		return &raw
	}
}
