// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type EntityState struct {
	LastChanged time.Time      `json:"last_changed"`
	LastUpdated time.Time      `json:"last_updated,omitempty"`
	State       any            `json:"state"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	EntityID    string         `json:"entity_id"`
	sensorType  string         `json:"-"`
}

func (e *EntityState) Auth() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.Token
}

func (e *EntityState) ResponseBody() any { return e }

func GetEntityState(ctx context.Context, sensorType, entityID string) (*EntityState, error) {
	prefs, err := preferences.Load()
	if err != nil {
		return nil, ErrNoPrefs
	}
	url := prefs.Host + "/api/states/" + sensorType + "." + prefs.DeviceName + "_" + entityID
	ctx = ContextSetURL(ctx, url)

	entity := &EntityState{
		EntityID:   entityID,
		sensorType: sensorType,
	}
	resp := <-ExecuteRequest(context.TODO(), entity)
	if resp.Error != nil {
		return nil, resp.Error
	}
	var e *EntityState
	var ok bool
	if e, ok = resp.Body.(*EntityState); !ok {
		return nil, ErrResponseMalformed
	}
	return e, nil
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

func (e *EntityStates) ResponseBody() any { return e }

func GetAllEntityStates() (*EntityStates, error) {
	resp := <-ExecuteRequest(context.TODO(), &EntityStates{})
	if resp.Error != nil {
		return nil, resp.Error
	}
	var entities *EntityStates
	var ok bool
	if entities, ok = resp.Body.(*EntityStates); !ok {
		return nil, ErrResponseMalformed
	}
	return entities, nil
}
