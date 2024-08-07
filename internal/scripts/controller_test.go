// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,wsl,godox,govet,containedctx,nlreturn
//revive:disable:unused-receiver
package scripts

import (
	"context"
	"log/slog"
	"reflect"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

func TestController_ActiveWorkers(t *testing.T) {
	type fields struct {
		scheduler *cron.Cron
		logger    *slog.Logger
		jobs      []job
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "with active jobs",
			fields: fields{
				jobs: []job{{ID: 1, Script: Script{path: "active"}}, {ID: 0, Script: Script{path: "inactive"}}, {}},
			},
			want: []string{"active"},
		},
		{
			name: "with no jobs",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			if got := c.ActiveWorkers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Controller.ActiveWorkers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_InactiveWorkers(t *testing.T) {
	type fields struct {
		scheduler *cron.Cron
		logger    *slog.Logger
		jobs      []job
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "with inactive jobs",
			fields: fields{
				jobs: []job{{ID: 1, Script: Script{path: "active"}}, {ID: 0, Script: Script{path: "inactive"}}, {}},
			},
			want: []string{"inactive", ""},
		},
		{
			name: "with no jobs",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			if got := c.InactiveWorkers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Controller.InactiveWorkers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_Start(t *testing.T) {
	type fields struct {
		scheduler *cron.Cron
		logger    *slog.Logger
		jobs      []job
	}
	type args struct {
		in0  context.Context
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    <-chan sensor.Details
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			got, err := c.Start(tt.args.in0, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Controller.Start() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_Stop(t *testing.T) {
	type fields struct {
		scheduler *cron.Cron
		logger    *slog.Logger
		jobs      []job
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			if err := c.Stop(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Controller.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_StartAll(t *testing.T) {
	type fields struct {
		scheduler *cron.Cron
		logger    *slog.Logger
		jobs      []job
	}
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    <-chan sensor.Details
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			got, err := c.StartAll(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.StartAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Controller.StartAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_StopAll(t *testing.T) {
	type fields struct {
		scheduler *cron.Cron
		logger    *slog.Logger
		jobs      []job
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			if err := c.StopAll(); (err != nil) != tt.wantErr {
				t.Errorf("Controller.StopAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewScriptsController(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Controller
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScriptsController(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScriptsController() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScriptsController() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findScripts(t *testing.T) {
	script, err := NewScript("testing/data/jsonTestScript.sh")
	require.NoError(t, err)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []Script
		wantErr bool
	}{
		{
			name: "with scripts",
			args: args{path: "testing/data"},
			want: []Script{*script},
		},
		{
			name: "without scripts",
			args: args{path: "foo/bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findScripts(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("findScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findScripts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isExecutable(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is executable",
			args: args{filename: "/proc/self/exe"},
			want: true,
		},
		{
			name: "is not executable",
			args: args{filename: "/does/not/exist"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExecutable(tt.args.filename); got != tt.want {
				t.Errorf("isExecutable() = %v, want %v", got, tt.want)
			}
		})
	}
}
