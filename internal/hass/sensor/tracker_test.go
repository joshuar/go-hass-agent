// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,wsl,nlreturn
//revive:disable:unused-parameter
//go:generate moq -out tracker_mocks_test.go . Registry
package sensor

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var mockSensorMap = map[string]Details{
	sensorExistingID: mockSensorRegistration,
}

var mockTracker = &Tracker{
	sensor: mockSensorMap,
}

func TestTracker_Get(t *testing.T) {
	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	type args struct {
		id string
	}
	tests := []struct {
		want    Details
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "successful get",
			fields:  fields{sensor: mockSensorMap},
			args:    args{id: sensorExistingID},
			wantErr: false,
			want:    mockSensorRegistration,
		},
		{
			name:    "unsuccessful get",
			fields:  fields{sensor: mockSensorMap},
			args:    args{id: "doesntExist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			got, err := tr.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tracker.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tracker.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracker_SensorList(t *testing.T) {
	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "with sensors",
			fields: fields{sensor: mockSensorMap},
			want:   []string{sensorExistingID},
		},
		{
			name: "without sensors",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if got := tr.SensorList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tracker.SensorList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracker_Process(t *testing.T) {
	restAPIURL := "http://localhost:8123/api"
	prefs := preferences.DefaultPreferences()
	prefs.RestAPIURL = restAPIURL
	ctx := preferences.ContextSetPrefs(context.TODO(), prefs)
	ctx, err := hass.SetupContext(ctx)
	require.NoError(t, err)

	sensorCh := make(chan Details)
	go func() {
		sensorCh <- mockSensorRegistration
		close(sensorCh)
	}()

	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	type args struct {
		ctx  context.Context //nolint:containedctx
		reg  Registry
		upds []<-chan Details
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "with channel contents",
			args: args{ctx: ctx, reg: mockRegistry, upds: []<-chan Details{sensorCh}},
		},
		{
			name: "without channel contents",
			args: args{ctx: context.TODO()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if err := tr.Process(tt.args.ctx, tt.args.reg, tt.args.upds...); (err != nil) != tt.wantErr {
				t.Errorf("Tracker.Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTracker_add(t *testing.T) {
	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	type args struct {
		sensor Details
	}
	tests := []struct {
		args    args
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "successful add",
			fields: fields{
				sensor: make(map[string]Details),
			},
			args:    args{sensor: mockSensorRegistration},
			wantErr: false,
		},
		{
			name:    "unsuccessful add (not initialised properly)",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			if err := tr.add(tt.args.sensor); (err != nil) != tt.wantErr {
				t.Errorf("Tracker.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// update wraps functions tested elsewhere.
// func TestTracker_update(t *testing.T) {}

func Test_handleResponse(t *testing.T) {
	updateSensorResp := &updateResponse{
		Body: map[string]*response{sensorExistingID: {Success: true}},
	}
	updateSensor := *mockSensorRegistration
	updateSensor.IDFunc = func() string { return sensorExistingID }

	newSensorResp := &registrationResponse{Body: response{Success: true}}

	type args struct {
		respIntr hass.Response
		trk      *Tracker
		upd      Details
		reg      Registry
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "update sensor",
			args: args{respIntr: updateSensorResp, trk: mockTracker, reg: mockRegistry, upd: &updateSensor},
		},
		{
			name: "new sensor",
			args: args{respIntr: newSensorResp, trk: mockTracker, reg: mockRegistry, upd: mockSensorRegistration},
		},
		{
			name: "location",
			args: args{respIntr: &locationResponse{}, upd: mockSensorRegistration},
		},
		{
			name:    "invalid sensor",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleResponse(tt.args.respIntr, tt.args.trk, tt.args.upd, tt.args.reg); (err != nil) != tt.wantErr {
				t.Errorf("handleResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleUpdates(t *testing.T) {
	successResp := &updateResponse{
		Body: map[string]*response{sensorExistingID: {Success: true}},
	}
	apiFailResp := &updateResponse{
		Body:     map[string]*response{sensorExistingID: {Success: false}},
		APIError: &hass.APIError{Code: 404, Message: "not found"},
	}
	haFailResp := &updateResponse{
		Body: map[string]*response{sensorExistingID: {Success: false, Error: &haError{Code: 401, Message: "method not allowed"}}},
	}
	disabledResp := &updateResponse{
		Body: map[string]*response{sensorExistingID: {Success: true, Disabled: true}},
	}

	type args struct {
		reg Registry
		r   *updateResponse
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{reg: mockRegistry, r: successResp},
		},
		{
			name:    "api fail",
			args:    args{reg: mockRegistry, r: apiFailResp},
			wantErr: true,
		},
		{
			name:    "ha fail",
			args:    args{reg: mockRegistry, r: haFailResp},
			wantErr: true,
		},
		{
			name: "disabled",
			args: args{reg: mockRegistry, r: disabledResp},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleUpdates(tt.args.reg, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("handleUpdates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// handleRegistration is indirectly covered
// func Test_handleRegistration(t *testing.T) {}

func TestTracker_Reset(t *testing.T) {
	type fields struct {
		sensor map[string]Details
		mu     sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "default",
			fields: fields{sensor: mockSensorMap},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
				mu:     tt.fields.mu,
			}
			tr.Reset()
		})
	}
}

func TestNewTracker(t *testing.T) {
	tests := []struct {
		want    *Tracker
		name    string
		wantErr bool
	}{
		{
			name: "success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTracker()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTracker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// MergeSensorCh is indirectly tested.
// func TestMergeSensorCh(t *testing.T) {}
