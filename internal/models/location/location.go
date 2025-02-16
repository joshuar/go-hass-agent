// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package sensor provides a method and options for creating model.Location
// objects wrapped as a model.Entity.
package location

import (
	"context"
	"errors"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/models"
)

var ErrNewLocation = errors.New("could not create new location")

type Option models.Option[*models.Location]

// WithGPSCoords sets the latitude and longitude GPS coordinates for the location.
func WithGPSCoords(latitude float32, longitude float32) Option {
	return func(l *models.Location) error {
		l.Gps = []float32{latitude, longitude}
		return nil
	}
}

// WithGPSAccuracy option sets the GPS accuracy value for the location.
func WithGPSAccuracy(accuracy int) Option {
	return func(l *models.Location) error {
		l.GpsAccuracy = accuracy
		return nil
	}
}

// WithSpeed option sets the speed value for the location.
func WithSpeed(speed int) Option {
	return func(l *models.Location) error {
		l.Speed = &speed
		return nil
	}
}

// WithAltitude option sets the altitude value for the location.
func WithAltitude(altitude int) Option {
	return func(l *models.Location) error {
		l.Altitude = &altitude
		return nil
	}
}

// NewLocation provides a way to build a location entity with the given options.
func NewLocation(ctx context.Context, options ...Option) (models.Entity, error) {
	location := models.Location{}

	for _, option := range options {
		if err := option(&location); err != nil {
			logging.FromContext(ctx).Warn("Could not set location option.", slog.Any("error", err))
		}
	}

	entity := models.Entity{}

	err := entity.FromLocation(location)
	if err != nil {
		return entity, errors.Join(ErrNewLocation, err)
	}

	return entity, nil
}
