// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"encoding/json"

	"github.com/joshuar/go-hass-agent/internal/request"
	"github.com/rs/zerolog/log"
)

// LocationUpdate represents the location information that can be sent to HA
// to update the location of the agent.
type LocationUpdate struct {
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

func (l *LocationUpdate) RequestType() request.RequestType {
	return request.RequestTypeUpdateLocation
}

func (l *LocationUpdate) RequestData() json.RawMessage {
	data, _ := json.Marshal(l)
	raw := json.RawMessage(data)
	return raw
}

func (l *LocationUpdate) ResponseHandler(resp bytes.Buffer) {
	if resp.Len() == 0 {
		log.Debug().Caller().Msg("No response data.")
	} else {
		log.Debug().Caller().Msg("Location updated.")
	}
}
