// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package location

import (
	"context"
	"errors"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/components/validation"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/models"
)

var ErrHandleLocation = errors.New("error handling location data")

type API interface {
	SendRequest(ctx context.Context, url string, req api.Request) (api.Response, error)
}

// newLocationRequest takes location data and creates a location request.
func newLocationRequest(location *models.Location) (*api.Request, error) {
	if valid, problems := validation.ValidateStruct(location); !valid {
		return nil, errors.Join(ErrHandleLocation, problems)
	}

	req := &api.Request{
		Type: api.UpdateLocation,
	}

	// Add the sensor registration into the request.
	err := req.Data.FromLocation(*location)
	if err != nil {
		return nil, errors.Join(ErrHandleLocation, err)
	}

	return req, nil
}

// Handler handles sending location data as a request to Home Assistant and
// processing the response.
func Handler(ctx context.Context, client API, location models.Location) error {
	req, err := newLocationRequest(&location)
	if err != nil {
		return errors.Join(ErrHandleLocation, err)
	}

	resp, err := client.SendRequest(ctx, preferences.RestAPIURL(), *req)
	if err != nil {
		return errors.Join(ErrHandleLocation, err)
	}

	status, err := resp.AsResponseStatus()
	if err != nil {
		return errors.Join(ErrHandleLocation, err)
	}

	if err := status.HasError(); err != nil {
		return errors.Join(ErrHandleLocation, err)
	}

	logging.FromContext(ctx).Debug("Location sent.")

	return nil
}
