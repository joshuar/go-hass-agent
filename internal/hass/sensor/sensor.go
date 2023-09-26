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
	State             interface{} `json:"state"`
	StateAttributes   interface{} `json:"attributes,omitempty"`
	UniqueID          string      `json:"unique_id"`
	Type              string      `json:"type"`
	Name              string      `json:"name"`
	UnitOfMeasurement string      `json:"unit_of_measurement,omitempty"`
	StateClass        string      `json:"state_class,omitempty"`
	EntityCategory    string      `json:"entity_category,omitempty"`
	Icon              string      `json:"icon,omitempty"`
	DeviceClass       string      `json:"device_class,omitempty"`
	Disabled          bool        `json:"disabled,omitempty"`
}

func (reg *SensorRegistrationInfo) RequestType() api.RequestType {
	return api.RequestTypeRegisterSensor
}

func (reg *SensorRegistrationInfo) RequestData() json.RawMessage {
	data, err := json.Marshal(reg)
	if err != nil {
		log.Debug().Err(err).
			Msg("Unable to marshal sensor to json.")
		return nil
	}
	return data
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

func (upd *SensorUpdateInfo) RequestType() api.RequestType {
	return api.RequestTypeUpdateSensorStates
}

func (upd *SensorUpdateInfo) RequestData() json.RawMessage {
	data, err := json.Marshal(upd)
	if err != nil {
		log.Debug().Err(err).
			Msg("Unable to marshal sensor to json.")
		return nil
	}
	return data
}
