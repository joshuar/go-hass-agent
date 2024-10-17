// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package event

type Event struct {
	EventData any    `json:"event_data" validate:"required"`
	EventType string `json:"event_type" validate:"required"`
}
