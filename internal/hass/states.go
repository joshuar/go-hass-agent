// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type EntityState struct {
	LastChanged time.Time      `json:"last_changed"`
	LastUpdated time.Time      `json:"last_updated,omitempty"`
	State       any            `json:"state"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	EntityID    string         `json:"entity_id"`
}

type EntityStateRequest struct{}

func (e *EntityStateRequest) Auth() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.Token
}

type EntityStateResponse struct {
	State *EntityState
}

func (e *EntityStateResponse) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, e.State)
}

func GetEntityState(sensorType, entityID string) (*EntityStateResponse, error) {
	ctx := context.TODO()
	prefs, err := preferences.Load()
	if err != nil {
		return nil, ErrNoPrefs
	}
	url := prefs.Host + "/api/states/" + sensorType + "." + prefs.DeviceName + "_" + entityID
	ctx = ContextSetURL(ctx, url)
	ctx = ContextSetClient(ctx, resty.New())

	req := &EntityStateRequest{}
	resp := &EntityStateResponse{}
	if err := ExecuteRequest(ctx, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

type EntityStatesRequest struct{}

type EntityStatesResponse struct {
	States []EntityStateResponse
}

func (e *EntityStatesResponse) Auth() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.Token
}

func (e *EntityStatesResponse) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &e.States)
}

func GetAllEntityStates() (*EntityStatesResponse, error) {
	ctx := context.TODO()
	prefs, err := preferences.Load()
	if err != nil {
		return nil, ErrNoPrefs
	}
	url := prefs.Host + "/api/states/"
	ctx = ContextSetURL(ctx, url)
	ctx = ContextSetClient(ctx, resty.New())

	req := &EntityStatesRequest{}
	resp := &EntityStatesResponse{}
	if err := ExecuteRequest(ctx, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
