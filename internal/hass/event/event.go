// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package event contains code for processing events from workers through the Home Assistant API.
package event

import (
	"context"
	"fmt"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/components/validation"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/models"
)

type clientAPI interface {
	SendRequest(ctx context.Context, url string, req api.Request) (api.Response, error)
}

// NewEventRequest takes event data and creates an event request.
func newEventRequest(event *models.Event) (*api.Request, error) {
	if valid, problems := validation.ValidateStruct(event); !valid {
		return nil, fmt.Errorf("could not marshal event data: %w", problems)
	}

	req := &api.Request{
		Type:      api.FireEvent,
		Data:      &api.Request_Data{},
		Retryable: event.Retryable,
	}

	// Add the sensor registration into the request.
	err := req.Data.FromEvent(*event)
	if err != nil {
		return nil, fmt.Errorf("could not marshal event data: %w", err)
	}

	return req, nil
}

// Handler handles sending event data as a request to Home Assistant and
// processing the response.
func Handler(ctx context.Context, client clientAPI, event models.Event) error {
	req, err := newEventRequest(&event)
	if err != nil {
		return err
	}

	resp, err := client.SendRequest(ctx, preferences.RestAPIURL(), *req)
	if err != nil {
		return fmt.Errorf("could not send event data: %w", err)
	}

	status, err := resp.AsResponseStatus()
	if err != nil {
		return fmt.Errorf("could not marshal event response: %w", err)
	}

	if err := status.HasError(); err != nil {
		return fmt.Errorf("could not determine response status: %w", err)
	}

	slogctx.FromCtx(ctx).Debug("Event sent.", event.LogAttributes())

	return nil
}
