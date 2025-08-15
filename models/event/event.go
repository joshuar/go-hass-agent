// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package event provides a method and options for creating an event entity.
package event

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/models"
)

// NewEvent creates an event entity with the given options.
func NewEvent(eventType string, eventData map[string]any) (models.Entity, error) {
	event := models.Event{
		Type: eventType,
		Data: eventData,
	}

	entity := models.Entity{}

	err := entity.FromEvent(event)
	if err != nil {
		return entity, fmt.Errorf("could not generate event entity: %w", err)
	}

	return entity, nil
}
