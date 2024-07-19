// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,wsl,containedctx,nlreturn
package agent

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/scripts"
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

//nolint:containedctx
func TestFindScripts(t *testing.T) {
	script, err := scripts.NewScript("testing/data/jsonTestScript.sh")
	require.NoError(t, err)

	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []Script
		wantErr bool
	}{
		{
			name: "path with scripts",
			args: args{ctx: context.TODO(), path: "testing/data"},
			want: []Script{script},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findScripts(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got[0], tt.want[0]) {
				t.Errorf("FindScripts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_runScripts(t *testing.T) {
	validScript := &ScriptMock{
		ScheduleFunc: func() string { return "@very 5s" },
	}

	goodTracker := &SensorTrackerMock{
		ProcessFunc: func(_ context.Context, _ sensor.Registry, _ ...<-chan sensor.Details) error {
			return nil
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
		ctx           context.Context
		trk           SensorTracker
		reg           sensor.Registry
		sensorScripts []Script
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "without scripts",
			fields: fields{logger: slog.Default()},
		},
		{
			name:   "with scripts",
			args:   args{sensorScripts: []Script{validScript}, trk: goodTracker},
			fields: fields{logger: slog.Default()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			agent.runScripts(tt.args.ctx, tt.args.trk, tt.args.reg, tt.args.sensorScripts...)
		})
	}
}
