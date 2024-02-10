// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type EntityState struct {
	LastChanged time.Time      `json:"last_changed"`
	LastUpdated time.Time      `json:"last_updated,omitifempty"`
	State       any            `json:"state"`
	Attributes  map[string]any `json:"attributes,omitifempty"`
	EntityID    string         `json:"entity_id"`
	sensorType  string         `json:"-"`
}

func (e *EntityState) URL() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.Host + "/api/states/" + e.sensorType + "." + prefs.DeviceName + "_" + e.EntityID
}

func (e *EntityState) Auth() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.Token
}

func (e *EntityState) Body() json.RawMessage {
	return nil
}

func GetEntityState(sensorType, entityID string) (*EntityState, error) {
	entity := &EntityState{
		EntityID:   entityID,
		sensorType: sensorType,
	}
	resp := <-api.ExecuteRequest2(context.TODO(), entity)
	if resp.Error != nil {
		return nil, resp.Error
	}
	err := json.Unmarshal(resp.Body, &entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

type EntityStates []EntityState

func (e *EntityStates) URL() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.Host + "/api/states"
}

func (e *EntityStates) Auth() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.Token
}

func (e *EntityStates) Body() json.RawMessage {
	return nil
}

func GetAllEntityStates() (*EntityStates, error) {
	entities := &EntityStates{}
	resp := <-api.ExecuteRequest2(context.TODO(), entities)
	if resp.Error != nil {
		return nil, resp.Error
	}
	err := json.Unmarshal(resp.Body, &entities)
	if err != nil {
		return nil, err
	}
	return entities, nil
}
