// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package location contains code for processing location requests through the Home Assistant API.
package location

import (
	"context"
	"fmt"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/hass/api"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/validation"
)

type clientAPI interface {
	SendRequest(ctx context.Context, url string, req api.RequestData) (api.ResponseData, error)
	RestAPIURL() string
}

// newLocationRequest takes location data and creates a location request.
func newLocationRequest(location *models.Location) (*api.RequestData, error) {
	if valid, problems := validation.ValidateStruct(location); !valid {
		return nil, fmt.Errorf("could not marshal location data: %w", problems)
	}

	req := &api.RequestData{
		Type: api.UpdateLocation,
	}

	// Add the sensor registration into the request.
	err := req.Payload.FromLocation(*location)
	if err != nil {
		return nil, fmt.Errorf("could not marshal location data: %w", err)
	}

	return req, nil
}

// Handler handles sending location data as a request to Home Assistant and
// processing the response.
func Handler(ctx context.Context, client clientAPI, location models.Location) error {
	req, err := newLocationRequest(&location)
	if err != nil {
		return err
	}

	resp, err := client.SendRequest(ctx, client.RestAPIURL(), *req)
	if err != nil {
		return fmt.Errorf("could not send location request: %w", err)
	}

	status, err := resp.AsResponseStatus()
	if err != nil {
		return fmt.Errorf("could not marshal location response: %w", err)
	}

	if err := status.HasError(); err != nil {
		return fmt.Errorf("could not determine location response status: %w", err)
	}

	slogctx.FromCtx(ctx).Debug("Location sent.")

	return nil
}
