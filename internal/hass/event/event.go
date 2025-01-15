// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package event

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/components/validation"
)

const (
	requestTypeEvent = "fire_event"
)

type Event struct {
	EventData    any    `json:"event_data" validate:"required"`
	EventType    string `json:"event_type" validate:"required"`
	RetryRequest bool
}

func (e *Event) Validate() error {
	err := validation.Validate.Struct(e)
	if err != nil {
		return fmt.Errorf("event is invalid: %s", validation.ParseValidationErrors(err))
	}

	return nil
}

func (e *Event) RequestBody() any {
	return struct {
		Data        any    `json:"data"`
		RequestType string `json:"type"`
	}{
		RequestType: requestTypeEvent,
		Data:        e,
	}
}

func (e *Event) Retry() bool {
	return e.RetryRequest
}
