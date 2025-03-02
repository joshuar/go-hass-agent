// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package event

import (
	"context"
	"errors"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/components/validation"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/models"
)

var ErrHandleEvent = errors.New("error handling event data")

type API interface {
	SendRequest(ctx context.Context, url string, req api.Request) (api.Response, error)
}

// NewEventRequest takes event data and creates an event request.
func newEventRequest(event *models.Event) (*api.Request, error) {
	if valid, problems := validation.ValidateStruct(event); !valid {
		return nil, errors.Join(ErrHandleEvent, problems)
	}

	req := &api.Request{
		Type:      api.FireEvent,
		Data:      &api.Request_Data{},
		Retryable: event.Retryable,
	}

	// Add the sensor registration into the request.
	err := req.Data.FromEvent(*event)
	if err != nil {
		return nil, errors.Join(ErrHandleEvent, err)
	}

	return req, nil
}

// Handler handles sending event data as a request to Home Assistant and
// processing the response.
func Handler(ctx context.Context, client API, event models.Event) error {
	req, err := newEventRequest(&event)
	if err != nil {
		return errors.Join(ErrHandleEvent, err)
	}

	resp, err := client.SendRequest(ctx, preferences.RestAPIURL(), *req)
	if err != nil {
		return errors.Join(ErrHandleEvent, err)
	}

	status, err := resp.AsResponseStatus()
	if err != nil {
		return errors.Join(ErrHandleEvent, err)
	}

	if err := status.HasError(); err != nil {
		return errors.Join(ErrHandleEvent, err)
	}

	logging.FromContext(ctx).Debug("Event sent.", event.LogAttributes())

	return nil
}
