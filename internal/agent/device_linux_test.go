// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,varnamelen,wsl,nlreturn,err113
package agent

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"

	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/stretchr/testify/assert"

	mocks "github.com/joshuar/go-hass-agent/internal/agent/testing"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

func Test_linuxController_ActiveWorkers(t *testing.T) {
	activeWorker := &sensorWorker{started: true}

	type fields struct {
		sensorWorkers map[string]*sensorWorker
		dbusAPI       *dbusx.DBusAPI
		logger        *slog.Logger
		mqttDevice    *mqtthass.Device
		mqttWorker    *mqttWorker
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "valid",
			fields: fields{sensorWorkers: map[string]*sensorWorker{"active_worker": activeWorker}},
			want:   []string{"active_worker"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := linuxController{
				sensorWorkers: tt.fields.sensorWorkers,
				dbusAPI:       tt.fields.dbusAPI,
				logger:        tt.fields.logger,
				mqttDevice:    tt.fields.mqttDevice,
				mqttWorker:    tt.fields.mqttWorker,
			}
			if got := w.ActiveWorkers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("linuxController.ActiveWorkers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linuxController_InactiveWorkers(t *testing.T) {
	inactiveWorker := &sensorWorker{}

	type fields struct {
		sensorWorkers map[string]*sensorWorker
		dbusAPI       *dbusx.DBusAPI
		logger        *slog.Logger
		mqttDevice    *mqtthass.Device
		mqttWorker    *mqttWorker
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "valid",
			fields: fields{sensorWorkers: map[string]*sensorWorker{"inactive_worker": inactiveWorker}},
			want:   []string{"inactive_worker"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := linuxController{
				sensorWorkers: tt.fields.sensorWorkers,
				dbusAPI:       tt.fields.dbusAPI,
				logger:        tt.fields.logger,
				mqttDevice:    tt.fields.mqttDevice,
				mqttWorker:    tt.fields.mqttWorker,
			}
			if got := w.InactiveWorkers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("linuxController.InactiveWorkers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linuxController_Start(t *testing.T) {
	outCh := make(<-chan sensor.Details)

	worker := &mocks.MockWorker{}
	worker.EXPECT().Updates(context.TODO()).Return(outCh, nil)

	type fields struct {
		sensorWorkers map[string]*sensorWorker
		dbusAPI       *dbusx.DBusAPI
		logger        *slog.Logger
		mqttDevice    *mqtthass.Device
		mqttWorker    *mqttWorker
	}
	type args struct {
		name string
	}
	tests := []struct {
		fields       fields
		wantErrValue error
		want         <-chan sensor.Details
		args         args
		name         string
		wantErr      bool
	}{
		{
			name:         "valid",
			args:         args{name: "valid"},
			fields:       fields{sensorWorkers: map[string]*sensorWorker{"valid": {object: worker}}},
			want:         outCh,
			wantErr:      false,
			wantErrValue: nil,
		},
		{
			name:         "unknown worker",
			args:         args{name: "unknown"},
			wantErr:      true,
			wantErrValue: ErrUnknownWorker,
		},
		{
			name:         "already started",
			args:         args{name: "started"},
			fields:       fields{sensorWorkers: map[string]*sensorWorker{"started": {object: worker, started: true}}},
			wantErr:      true,
			wantErrValue: ErrWorkerAlreadyStarted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := linuxController{
				sensorWorkers: tt.fields.sensorWorkers,
				dbusAPI:       tt.fields.dbusAPI,
				logger:        tt.fields.logger,
				mqttDevice:    tt.fields.mqttDevice,
				mqttWorker:    tt.fields.mqttWorker,
			}
			got, err := w.Start(context.TODO(), tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("linuxController.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("linuxController.Start() = %v, want %v", got, tt.want)
			}
			assert.ErrorIs(t, err, tt.wantErrValue)
		})
	}
}

func Test_linuxController_Stop(t *testing.T) {
	goodWorker := &mocks.MockWorker{}
	goodWorker.EXPECT().Stop().Return(nil)
	badWorker := &mocks.MockWorker{}
	badWorker.EXPECT().Stop().Return(errors.New("i did not stop"))

	type fields struct {
		sensorWorkers map[string]*sensorWorker
		dbusAPI       *dbusx.DBusAPI
		logger        *slog.Logger
		mqttDevice    *mqtthass.Device
		mqttWorker    *mqttWorker
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "successful",
			args:    args{name: "good"},
			fields:  fields{sensorWorkers: map[string]*sensorWorker{"good": {object: goodWorker, started: true}}},
			wantErr: false,
		},
		{
			name:    "unsuccessful",
			args:    args{name: "bad"},
			fields:  fields{sensorWorkers: map[string]*sensorWorker{"bad": {object: badWorker, started: true}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := linuxController{
				sensorWorkers: tt.fields.sensorWorkers,
				dbusAPI:       tt.fields.dbusAPI,
				logger:        tt.fields.logger,
				mqttDevice:    tt.fields.mqttDevice,
				mqttWorker:    tt.fields.mqttWorker,
			}
			if err := w.Stop(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("linuxController.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
