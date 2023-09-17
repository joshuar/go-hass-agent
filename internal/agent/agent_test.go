// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	_ "embed"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/tracker"
)

func Test_newAgent(t *testing.T) {
	type args struct {
		appID    string
		headless bool
	}
	tests := []struct {
		name string
		args args
		want *Agent
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newAgent(tt.args.appID, tt.args.headless); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	type args struct {
		options AgentOptions
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Run(tt.args.options)
		})
	}
}

func TestRegister(t *testing.T) {
	type args struct {
		options AgentOptions
		server  string
		token   string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Register(tt.args.options, tt.args.server, tt.args.token)
		})
	}
}

func TestShowVersion(t *testing.T) {
	type args struct {
		options AgentOptions
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ShowVersion(tt.args.options)
		})
	}
}

func TestShowInfo(t *testing.T) {
	type args struct {
		options AgentOptions
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ShowInfo(tt.args.options)
		})
	}
}

func TestAgent_setupLogging(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			agent.setupLogging(tt.args.ctx)
		})
	}
}

func TestAgent_handleSignals(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	type args struct {
		cancelFunc context.CancelFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			agent.handleSignals(tt.args.cancelFunc)
		})
	}
}

func TestAgent_handleShutdown(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			agent.handleShutdown(tt.args.ctx)
		})
	}
}

func TestAgent_AppName(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			if got := agent.AppName(); got != tt.want {
				t.Errorf("Agent.AppName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_AppID(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			if got := agent.AppID(); got != tt.want {
				t.Errorf("Agent.AppID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_AppVersion(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			if got := agent.AppVersion(); got != tt.want {
				t.Errorf("Agent.AppVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_Stop(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			agent.Stop()
		})
	}
}

func TestAgent_GetConfig(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	type args struct {
		key   string
		value interface{}
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
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			if err := agent.GetConfig(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Agent.GetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_SetConfig(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	type args struct {
		key   string
		value interface{}
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
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			if err := agent.SetConfig(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Agent.SetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_StoragePath(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			got, err := agent.StoragePath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Agent.StoragePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Agent.StoragePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_SensorList(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			if got := agent.SensorList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Agent.SensorList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_SensorValue(t *testing.T) {
	type fields struct {
		ui      AgentUI
		config  AgentConfig
		sensors *tracker.SensorTracker
		done    chan struct{}
		name    string
		id      string
		version string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    tracker.Sensor
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				config:  tt.fields.config,
				sensors: tt.fields.sensors,
				done:    tt.fields.done,
				name:    tt.fields.name,
				id:      tt.fields.id,
				version: tt.fields.version,
			}
			got, err := agent.SensorValue(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Agent.SensorValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Agent.SensorValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
