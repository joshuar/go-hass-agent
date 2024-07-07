// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:errname // structs are dual-purpose response and error
//revive:disable:unused-receiver
package hass

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	statesEndpoint = "/api/states/"
)

type EntityState struct {
	LastChanged time.Time      `json:"last_changed"`
	LastUpdated time.Time      `json:"last_updated,omitempty"`
	State       any            `json:"state"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	EntityID    string         `json:"entity_id"`
}

type EntityStateRequest struct {
	token string
}

func (e *EntityStateRequest) Auth() string {
	return e.token
}

type EntityStateResponse struct {
	State *EntityState
	*APIError
}

func (e *EntityStateResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, e.State)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (e *EntityStateResponse) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, e.APIError)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (e *EntityStateResponse) Error() string {
	return e.APIError.Error()
}

//nolint:exhaustruct
func GetEntityState(ctx context.Context, sensorType, entityID string) (*EntityStateResponse, error) {
	prefs, err := preferences.ContextGetPrefs(ctx)
	if err != nil {
		return nil, ErrLoadPrefsFailed
	}

	client := NewDefaultHTTPClient(prefs.Host)
	url := statesEndpoint + sensorType + "." + prefs.DeviceName + "_" + entityID

	req := &EntityStateRequest{token: prefs.Token}
	resp := &EntityStateResponse{}

	if err := ExecuteRequest(ctx, client, url, req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type EntityStatesRequest struct {
	token string
}

func (e *EntityStatesRequest) Auth() string {
	return e.token
}

type EntityStatesResponse struct {
	*APIError
	States []EntityStateResponse
}

func (e *EntityStatesResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &e.States)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (e *EntityStatesResponse) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, e.APIError)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (e *EntityStatesResponse) Error() string {
	return e.APIError.Error()
}

//nolint:exhaustruct
func GetAllEntityStates(ctx context.Context) (*EntityStatesResponse, error) {
	prefs, err := preferences.ContextGetPrefs(ctx)
	if err != nil {
		return nil, ErrLoadPrefsFailed
	}

	client := NewDefaultHTTPClient(prefs.Host)
	req := &EntityStatesRequest{token: prefs.Token}
	resp := &EntityStatesResponse{}

	if err := ExecuteRequest(ctx, client, statesEndpoint, req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
