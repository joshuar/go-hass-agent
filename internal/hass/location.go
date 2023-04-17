// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"

	"github.com/rs/zerolog/log"
)

// LocationUpdate represents a location update from a platform/device. It
// provides a bridge between the platform/device specific location info and Home
// Assistant.
type LocationUpdate interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

// MarshalLocationUpdate will take a device type that implements LocationUpdate
// and marshal it into a locationUpdateInfo struct that can be sent as a request
// to HA
func MarshalLocationUpdate(l LocationUpdate) *locationUpdateInfo {
	return &locationUpdateInfo{
		Gps:              l.Gps(),
		GpsAccuracy:      l.GpsAccuracy(),
		Battery:          l.Battery(),
		Speed:            l.Speed(),
		Altitude:         l.Altitude(),
		Course:           l.Course(),
		VerticalAccuracy: l.VerticalAccuracy(),
	}
}

// locationUpdateInfo represents the location information that can be sent to HA
// to update the location of the agent.
type locationUpdateInfo struct {
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

func (l *locationUpdateInfo) RequestType() RequestType {
	return RequestTypeUpdateLocation
}

func (l *locationUpdateInfo) RequestData() interface{} {
	return l
}

func (l *locationUpdateInfo) ResponseHandler(resp bytes.Buffer) {
	if resp.Len() == 0 {
		log.Debug().Caller().Msg("No response data.")
	} else {
		log.Debug().Caller().Msg("Location updated.")
	}
}
