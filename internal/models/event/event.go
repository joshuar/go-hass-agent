// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package event

import (
	"errors"

	"github.com/joshuar/go-hass-agent/internal/models"
)

var ErrNewEvent = errors.New("could not create new event")

// NewEvent creates an event entity with the given options.
func NewEvent(eventType string, eventData map[string]any) (models.Entity, error) {
	event := models.Event{
		Type: eventType,
		Data: eventData,
	}

	entity := models.Entity{}

	err := entity.FromEvent(event)
	if err != nil {
		return entity, errors.Join(ErrNewEvent, err)
	}

	return entity, nil
}
