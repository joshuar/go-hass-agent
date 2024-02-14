// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
)

const (
	requestTypeLocation = "update_location"
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

type locationRequest struct {
	Data     *LocationData  `json:"data"`
	Response map[string]any `json:"-"`
	Type     string         `json:"type"`
}

func (l *locationRequest) RequestBody() json.RawMessage {
	data, err := json.Marshal(l)
	if err != nil {
		return nil
	}
	return json.RawMessage(data)
}

func (l *locationRequest) ResponseBody() any { return l.Response }

func UpdateLocation(ctx context.Context, l *LocationData) error {
	req := &locationRequest{
		Type:     requestTypeLocation,
		Data:     l,
		Response: make(map[string]any),
	}
	resp := <-ExecuteRequest(ctx, req)
	if resp.Error != nil {
		return resp.Error
	}
	log.Debug().Msg("location updated")
	return nil
}
