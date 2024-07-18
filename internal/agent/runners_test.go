// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,wsl,containedctx
package agent

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

//revive:disable:function-length
func TestAgent_runWorkers(t *testing.T) {
	sensorCh := make(chan sensor.Details)
	defer close(sensorCh)

	testCtx, testCancelFunc := context.WithCancel(context.TODO())
	defer testCancelFunc()

	registry := &RegistryMock{}

	errBadController := errors.New("bad controller") //nolint:err113
	errBadTracker := errors.New("bad tracker")       //nolint:err113

	goodController := &SensorControllerMock{
		StartAllFunc: func(_ context.Context) (<-chan sensor.Details, error) {
			return sensorCh, nil
		},
		StopAllFunc: func() error {
			return nil
		},
	}
	goodTracker := &SensorTrackerMock{
		ProcessFunc: func(_ context.Context, _ sensor.Registry, _ ...<-chan sensor.Details) error {
			return nil
		},
	}
	badController := &SensorControllerMock{
		StartAllFunc: func(_ context.Context) (<-chan sensor.Details, error) {
			return nil, errBadController
		},
	}
	badTracker := &SensorTrackerMock{
		ProcessFunc: func(_ context.Context, _ sensor.Registry, _ ...<-chan sensor.Details) error {
			return errBadTracker
		},
	}

	type fields struct {
		ui               UI
		done             chan struct{}
		registrationInfo *hass.RegistrationInput
		// prefs            *preferences.Preferences
		logger        *slog.Logger
		id            string
		headless      bool
		forceRegister bool
	}
	type args struct {
		ctx         context.Context
		trk         SensorTracker
		reg         sensor.Registry
		controllers []SensorController
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "bad controller",
			args:   args{ctx: testCtx, trk: badTracker, reg: registry, controllers: []SensorController{badController}},
			fields: fields{logger: slog.Default(), id: "go-hass-agent-test"},
		},
		{
			name:   "bad tracker",
			args:   args{ctx: testCtx, trk: badTracker, reg: registry, controllers: []SensorController{goodController}},
			fields: fields{logger: slog.Default(), id: "go-hass-agent-test"},
		},
		{
			name:   "successful",
			args:   args{ctx: testCtx, trk: goodTracker, reg: registry, controllers: []SensorController{goodController}},
			fields: fields{logger: slog.Default(), id: "go-hass-agent-test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { //revive:disable:unused-parameter
			agent := &Agent{
				ui:               tt.fields.ui,
				done:             tt.fields.done,
				registrationInfo: tt.fields.registrationInfo,
				// prefs:            tt.fields.prefs,
				logger:        tt.fields.logger,
				id:            tt.fields.id,
				headless:      tt.fields.headless,
				forceRegister: tt.fields.forceRegister,
			}
			agent.runWorkers(testCtx, tt.args.trk, tt.args.reg, tt.args.controllers...)
		})
	}
}
