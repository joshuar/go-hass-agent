// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"encoding/json"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"
)

const (
	StateUnknown = "unknown"
)

// SensorRegistrationInfo is the JSON structure required to register a sensor
// with HA.
type SensorRegistrationInfo struct {
	Name              string `json:"name,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	StateClass        string `json:"state_class,omitempty"`
	EntityCategory    string `json:"entity_category,omitempty"`
	DeviceClass       string `json:"device_class,omitempty"`
}

// SensorUpdateInfo is the JSON structure required to update HA with the current
// sensor state.
type SensorUpdateInfo struct {
	StateAttributes interface{} `json:"attributes,omitempty"`
	State           interface{} `json:"state"`
	Icon            string      `json:"icon,omitempty"`
	Type            string      `json:"type"`
	UniqueID        string      `json:"unique_id"`
}

type SensorState struct {
	SensorUpdateInfo
	SensorRegistrationInfo
	Disabled   bool `json:"disabled,omitempty"`
	Registered bool `json:"-"`
}

func (s *SensorState) RequestType() api.RequestType {
	if s.Registered {
		return api.RequestTypeUpdateSensorStates
	} else {
		return api.RequestTypeRegisterSensor
	}
}

func (s *SensorState) RequestData() json.RawMessage {
	data, err := json.Marshal(s)
	if err != nil {
		log.Debug().Err(err).
			Msg("Unable to marshal sensor to json.")
		return nil
	}
	return data
}
