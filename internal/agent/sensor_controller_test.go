// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
//revive:disable:max-public-structs,unused-receiver,unused-parameter,function-length
package agent

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type mockSensor struct {
	state string
	id    string
}

func (s *mockSensor) Name() string { return "Mock Sensor" }

func (s *mockSensor) ID() string { return s.id }

func (s *mockSensor) Icon() string { return "mdi:test" }

func (s *mockSensor) SensorType() types.SensorClass { return types.Sensor }

func (s *mockSensor) DeviceClass() types.DeviceClass { return 0 }

func (s *mockSensor) StateClass() types.StateClass { return 0 }

func (s *mockSensor) State() any { return s.state }

func (s *mockSensor) Units() string { return "" }

func (s *mockSensor) Category() string { return sensor.CategoryDiagnostic }

func (s *mockSensor) Attributes() map[string]any { return nil }

func TestAgent_processResponse(t *testing.T) {
	successfulLocationUpdate := &LocationUpdateResponseMock{
		UpdatedFunc: func() bool { return true },
	}

	type fields struct {
		ui            UI
		done          chan struct{}
		prefs         *preferences.Preferences
		logger        *slog.Logger
		id            string
		headless      bool
		forceRegister bool
	}
	type args struct {
		upd  sensor.Details
		resp any
		reg  Registry
		trk  Tracker
	}
	tests := []struct {
		args   args
		name   string
		fields fields
	}{
		{
			name:   "successful location update",
			args:   args{resp: successfulLocationUpdate, upd: &mockSensor{id: "mock_location"}},
			fields: fields{logger: slog.Default(), id: "go-hass-agent-test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:            tt.fields.ui,
				done:          tt.fields.done,
				prefs:         tt.fields.prefs,
				logger:        tt.fields.logger,
				id:            tt.fields.id,
				headless:      tt.fields.headless,
				forceRegister: tt.fields.forceRegister,
			}
			agent.processResponse(context.TODO(), tt.args.upd, tt.args.resp, tt.args.reg, tt.args.trk)
		})
	}
}

func Test_processStateUpdates(t *testing.T) {
	reg := &RegistryMock{
		IsDisabledFunc: func(id string) bool {
			switch id {
			case "disabled":
				return true
			default:
				return false
			}
		},
		SetDisabledFunc: func(id string, _ bool) error {
			if id != "disabledfail" {
				return nil
			}
			return ErrRegDisableFailed
		},
	}

	trk := &TrackerMock{
		AddFunc: func(details sensor.Details) error {
			switch details.ID() {
			case "addfailed":
				return ErrTrkUpdateFailed
			default:
				return nil
			}
		},
	}

	type args struct {
		trk    Tracker
		reg    Registry
		upd    sensor.Details
		status *sensor.UpdateStatus
	}
	tests := []struct {
		args        args
		wantErrType error
		name        string
		want        bool
		wantErr     bool
	}{
		{
			name: "successful update",
			args: args{trk: trk, reg: reg, upd: &mockSensor{id: "success"}, status: &sensor.UpdateStatus{Success: true}},
			want: true,
		},
		{
			name:        "unsuccessful update",
			args:        args{trk: trk, reg: reg, upd: &mockSensor{}, status: &sensor.UpdateStatus{Success: false}},
			want:        false,
			wantErr:     true,
			wantErrType: ErrStateUpdateFailed,
		},
		{
			name: "successful update, disabled sensor",
			args: args{trk: trk, reg: reg, upd: &mockSensor{id: "notdisabled"}, status: &sensor.UpdateStatus{Success: true, Disabled: true}},
			want: true,
		},
		{
			name:        "successful update, disabled sensor, disable failed",
			args:        args{trk: trk, reg: reg, upd: &mockSensor{id: "disabledfail"}, status: &sensor.UpdateStatus{Success: true, Disabled: true}},
			want:        true,
			wantErr:     true,
			wantErrType: ErrRegDisableFailed,
		},
		{
			name:        "successful update, tracker update failed",
			args:        args{trk: trk, reg: reg, upd: &mockSensor{id: "addfailed"}, status: &sensor.UpdateStatus{Success: true}},
			want:        true,
			wantErr:     true,
			wantErrType: ErrTrkUpdateFailed,
		},
		{
			name:        "no status",
			want:        false,
			wantErr:     true,
			wantErrType: ErrStateUpdateUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processStateUpdates(tt.args.trk, tt.args.reg, tt.args.upd, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("processStateUpdates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processStateUpdates() = %v, want %v", got, tt.want)
			}
			if tt.wantErr {
				assert.ErrorIs(t, err, tt.wantErrType)
			}
		})
	}
}

func Test_processRegistration(t *testing.T) {
	reg := &RegistryMock{
		SetRegisteredFunc: func(id string, _ bool) error {
			if id == "registeredfail" {
				return ErrRegistrationFailed
			}
			return nil
		},
	}

	trk := &TrackerMock{
		AddFunc: func(details sensor.Details) error {
			switch details.ID() {
			case "addfailed":
				return ErrTrkUpdateFailed
			default:
				return nil
			}
		},
	}

	success := &SensorRegistrationResponseMock{
		RegisteredFunc: func() bool { return true },
	}

	fail := &SensorRegistrationResponseMock{
		RegisteredFunc: func() bool { return false },
		ErrorFunc:      func() string { return "failed" },
	}

	type args struct {
		trk     Tracker
		reg     Registry
		upd     sensor.Details
		details SensorRegistrationResponse
	}
	tests := []struct {
		args        args
		wantErrType error
		name        string
		want        bool
		wantErr     bool
	}{
		{
			name: "successful registration",
			args: args{trk: trk, reg: reg, upd: &mockSensor{id: "success"}, details: success},
			want: true,
		},
		{
			name:        "unsuccessful registration",
			args:        args{trk: trk, reg: reg, upd: &mockSensor{id: "success"}, details: fail},
			want:        false,
			wantErr:     true,
			wantErrType: ErrRegistrationFailed,
		},
		{
			name:        "successful registration, registry update failed",
			args:        args{trk: trk, reg: reg, upd: &mockSensor{id: "registeredfail"}, details: success},
			want:        true,
			wantErr:     true,
			wantErrType: ErrRegAddFailed,
		},
		{
			name:        "successful registration, tracker add failed",
			args:        args{trk: trk, reg: reg, upd: &mockSensor{id: "addfailed"}, details: success},
			want:        true,
			wantErr:     true,
			wantErrType: ErrTrkUpdateFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processRegistration(tt.args.trk, tt.args.reg, tt.args.upd, tt.args.details)
			if (err != nil) != tt.wantErr {
				t.Errorf("processRegistration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processRegistration() = %v, want %v", got, tt.want)
			}
			if tt.wantErr {
				assert.ErrorIs(t, err, tt.wantErrType)
			}
		})
	}
}
