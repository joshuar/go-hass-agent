// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/joshuar/go-hass-agent/validation"
)

// NewEvent creates an event entity with the given options.
func NewEvent(eventType string, eventData map[string]any) (Entity, error) {
	event := Event{
		Type: eventType,
		Data: eventData,
	}

	entity := Entity{}

	if err := entity.FromEvent(event); err != nil {
		return entity, fmt.Errorf("could not generate event entity: %w", err)
	}

	return entity, nil
}

// String returns a string representation of an event.
func (e *Event) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Event Type: %s", e.Type)
	fmt.Fprintf(&b, "Event Data: %v", e.Data)

	return b.String()
}

// LogAttributes returns an slog.Group of log attributes for an event entity.
func (e *Event) LogAttributes() slog.Attr {
	return slog.Group("event",
		slog.String("type", e.Type),
	)
}

// Valid returns whether the event data is valid.
func (e *Event) Valid() bool {
	if err := validation.ValidateStruct(e); err != nil {
		return false
	}

	return true
}
