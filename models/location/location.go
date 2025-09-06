// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package location provides a method and options for creating model.Location
// objects wrapped as a model.Entity.
package location

import (
	"context"

	"github.com/joshuar/go-hass-agent/models"
)

// Option is a functional option for a location.
type Option models.Option[*models.Location]

// WithGPSCoords sets the latitude and longitude GPS coordinates for the location.
func WithGPSCoords(latitude float32, longitude float32) Option {
	return func(l *models.Location) {
		l.Gps = []float32{latitude, longitude}
	}
}

// WithGPSAccuracy option sets the GPS accuracy value for the location.
func WithGPSAccuracy(accuracy int) Option {
	return func(l *models.Location) {
		l.GpsAccuracy = accuracy
	}
}

// WithSpeed option sets the speed value for the location.
func WithSpeed(speed int) Option {
	return func(l *models.Location) {
		l.Speed = speed
	}
}

// WithAltitude option sets the altitude value for the location.
func WithAltitude(altitude int) Option {
	return func(l *models.Location) {
		l.Altitude = altitude
	}
}

// NewLocation provides a way to build a location entity with the given options.
func NewLocation(_ context.Context, options ...Option) models.Entity {
	location := models.Location{}

	for _, option := range options {
		option(&location)
	}

	entity := models.Entity{}
	entity.FromLocation(location)
	return entity
}
