// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"errors"

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
	Data *LocationData `json:"data"`
	Type string        `json:"type"`
}

func (l *locationRequest) RequestBody() json.RawMessage {
	data, err := json.Marshal(l)
	if err != nil {
		return nil
	}
	return json.RawMessage(data)
}

type locationResponse struct {
	err error
}

func (l *locationResponse) UnmarshalJSON(b []byte) error {
	return nil
}

func (l *locationResponse) StoreError(e error) {
	l.err = e
}

func (l *locationResponse) Error() string {
	return l.err.Error()
}

func UpdateLocation(ctx context.Context, l *LocationData) error {
	req := &locationRequest{
		Type: requestTypeLocation,
		Data: l,
	}
	resp := &locationResponse{}
	ExecuteRequest(ctx, req, resp)
	if errors.Is(resp, &APIError{}) {
		return resp
	}
	log.Debug().Msg("Location updated")
	return nil
}
