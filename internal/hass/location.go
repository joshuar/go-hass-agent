// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/joshuar/go-hass-agent/internal/api"
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

func (l *LocationUpdate) RequestType() api.RequestType {
	return api.RequestTypeUpdateLocation
}

func (l *LocationUpdate) RequestData() json.RawMessage {
	data, _ := json.Marshal(l)
	raw := json.RawMessage(data)
	return raw
}

func (l *LocationUpdate) ResponseHandler(res bytes.Buffer, respCh chan api.Response) {
	response := new(locationResponse)
	if res.Len() == 0 {
		response.err = errors.New("no response data")
		respCh <- response
	} else {
		respCh <- response
	}
}

type locationResponse struct {
	err error
}

func (l locationResponse) Registered() bool {
	log.Debug().Msg("Registered should not be called for location response.")
	return false
}

func (l locationResponse) Disabled() bool {
	log.Debug().Msg("Disabled should not be called for location response.")
	return false
}

func (l locationResponse) Error() error {
	return l.err
}

func (l locationResponse) Type() api.RequestType {
	return api.RequestTypeUpdateLocation
}
