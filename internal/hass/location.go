// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"encoding/json"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

// LocationData represents the location information that can be sent to HA
// to update the location of the agent.
type LocationData struct {
	Gps              []float64 `json:"gps"`
	GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
	Battery          int       `json:"battery,omitempty"`
	Speed            int       `json:"speed,omitempty"`
	Altitude         int       `json:"altitude,omitempty"`
	Course           int       `json:"course,omitempty"`
	VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
}

// locationUpdateInfo implements hass.Request so it can be sent to HA as a
// request

func (l *LocationData) RequestType() api.RequestType {
	return api.RequestTypeUpdateLocation
}

func (l *LocationData) RequestData() json.RawMessage {
	data, err := json.Marshal(l)
	if err != nil {
		return nil
	}
	return json.RawMessage(data)
}
