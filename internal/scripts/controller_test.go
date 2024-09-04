// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
//revive:disable:unused-receiver
package scripts

import (
	"context"
	"log/slog"
	"reflect"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
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
		want    <-chan sensor.Details
		args    args
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "unknown script",
			args: args{name: "unknown"},
			fields: fields{
				jobs: []job{{Script: Script{path: "echo true"}}, {Script: Script{path: "echo false"}}},
			},
			wantErr: true,
		},
		{
			name: "already started",
			args: args{name: "echo already started"},
			fields: fields{
				jobs: []job{{ID: 1, Script: Script{path: "echo already started"}}, {Script: Script{path: "echo false"}}},
			},
			wantErr: true,
		},
		{
			name: "start",
			args: args{name: "echo start"},
			fields: fields{
				scheduler: cron.New(),
				jobs:      []job{{Script: Script{path: "echo start", schedule: "@every 1s"}}, {Script: Script{path: "echo false"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			_, err := c.Start(tt.args.in0, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotEmpty(t, len(c.ActiveWorkers())) //nolint:testifylint
			}
		})
	}
}

func TestController_Stop(t *testing.T) {
	scheduler := cron.New()
	id, err := scheduler.AddFunc("@every 5s", func() {})
	require.NoError(t, err)
	scheduler.Start()

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
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "unknown script",
			args: args{name: "unknown"},
			fields: fields{
				jobs: []job{{Script: Script{path: "echo true"}}, {Script: Script{path: "echo false"}}},
			},
			wantErr: true,
		},
		{
			name: "already started",
			args: args{name: "echo already stopped"},
			fields: fields{
				jobs: []job{{Script: Script{path: "echo already stopped"}}, {Script: Script{path: "echo false"}}},
			},
			wantErr: true,
		},
		{
			name: "stop",
			args: args{name: "echo stop"},
			fields: fields{
				scheduler: scheduler,
				jobs:      []job{{ID: id, Script: Script{path: "echo stop", schedule: "@every 1s"}}},
			},
		},
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
			if !tt.wantErr {
				assert.NotEmpty(t, len(c.InactiveWorkers())) //nolint:testifylint
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
		args    args
		want    <-chan sensor.Details
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "start",
			fields: fields{
				scheduler: cron.New(),
				logger:    slog.Default(),
				jobs:      []job{{Script: Script{path: "echo start", schedule: "@every 1s"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				scheduler: tt.fields.scheduler,
				logger:    tt.fields.logger,
				jobs:      tt.fields.jobs,
			}
			_, err := c.StartAll(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.StartAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotEmpty(t, len(c.ActiveWorkers())) //nolint:testifylint
			}
		})
	}
}

func TestController_StopAll(t *testing.T) {
	scheduler := cron.New()
	id, err := scheduler.AddFunc("@every 5s", func() {})
	require.NoError(t, err)
	scheduler.Start()

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
		{
			name: "stop",
			fields: fields{
				scheduler: scheduler,
				logger:    slog.Default(),
				jobs:      []job{{ID: id, Script: Script{path: "echo stop", schedule: "@every 1s"}}},
			},
		},
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
			if !tt.wantErr {
				assert.NotEmpty(t, len(c.InactiveWorkers())) //nolint:testifylint
			}
		})
	}
}

func TestNewScriptsController(t *testing.T) {
	ctx := logging.ToContext(context.TODO(), slog.Default())

	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		want      *Controller
		args      args
		name      string
		wantErr   bool
		wantEmpty bool
	}{
		{
			name: "valid path",
			args: args{ctx: ctx, path: "testing/data"},
		},
		{
			name:      "invalid path",
			args:      args{ctx: ctx, path: "foo/bar"},
			wantEmpty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScriptsController(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScriptsController() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantEmpty {
				assert.NotEmpty(t, got.jobs)
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
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findScripts(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("findScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				assert.Equal(t, script.path, got[0].path)
				assert.Equal(t, script.schedule, got[0].schedule)
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
