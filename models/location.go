// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

import (
	"context"

	"github.com/joshuar/go-hass-agent/validation"
)

// LocationOption is a functional option for a location.
type LocationOption Option[*Location]

// WithGPSCoords sets the latitude and longitude GPS coordinates for the location.
func WithGPSCoords(latitude float32, longitude float32) LocationOption {
	return func(l *Location) {
		l.Gps = []float32{latitude, longitude}
	}
}

// WithGPSAccuracy option sets the GPS accuracy value for the location.
func WithGPSAccuracy(accuracy int) LocationOption {
	return func(l *Location) {
		l.GpsAccuracy = accuracy
	}
}

// WithSpeed option sets the speed value for the location.
func WithSpeed(speed int) LocationOption {
	return func(l *Location) {
		l.Speed = speed
	}
}

// WithAltitude option sets the altitude value for the location.
func WithAltitude(altitude int) LocationOption {
	return func(l *Location) {
		l.Altitude = altitude
	}
}

// NewLocation provides a way to build a location entity with the given options.
func NewLocation(_ context.Context, options ...LocationOption) Entity {
	location := Location{}

	for _, option := range options {
		option(&location)
	}

	entity := Entity{}
	entity.FromLocation(location)
	return entity
}

// Valid returns whether the location data is valid.
func (e *Location) Valid() bool {
	if err := validation.ValidateStruct(e); err != nil {
		return false
	}

	return true
}
