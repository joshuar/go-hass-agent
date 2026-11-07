// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"
)

func newTestSensor(id string, callback func(args ...any) (json.RawMessage, error)) *mqtthass.SensorEntity {
	entity := mqtthass.NewSensorEntity().
		WithDetails(
			mqtthass.App("test"),
			mqtthass.Name(id),
			mqtthass.ID(id),
			mqtthass.DeviceInfo(&mqtthass.Device{}),
		)
	if callback != nil {
		return entity.WithState(mqtthass.StateCallback(callback))
	}

	return entity.WithState()
}

func TestAppendStates(t *testing.T) {
	withState := newTestSensor("with_state", func(_ ...any) (json.RawMessage, error) {
		return json.RawMessage(`{"value":42}`), nil
	})
	noCallback := newTestSensor("no_callback", nil)
	failing := newTestSensor("failing", func(_ ...any) (json.RawMessage, error) {
		return nil, errors.New("state unavailable")
	})

	existing := mqttapi.NewMsg("existing/topic", []byte("x"))
	msgs := appendStates(slog.New(slog.DiscardHandler), []*mqttapi.Msg{existing},
		[]*mqtthass.SensorEntity{withState, noCallback, failing})

	require.Len(t, msgs, 2, "entities without a state should be skipped")
	require.Equal(t, existing, msgs[0])
	require.Equal(t, withState.StateTopic, msgs[1].Topic)
	require.JSONEq(t, `{"value":42}`, string(msgs[1].Message))
}
