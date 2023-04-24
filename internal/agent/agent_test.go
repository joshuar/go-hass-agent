// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	_ "embed"
	"reflect"
	"sync"
	"testing"

	"fyne.io/fyne/v2"
)

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name string
		want *Agent
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAgent(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Run()
		})
	}
}

func TestAgent_stop(t *testing.T) {
	type fields struct {
		App     fyne.App
		Tray    fyne.Window
		Name    string
		Version string
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
				app:     tt.fields.App,
				tray:    tt.fields.Tray,
				Name:    tt.fields.Name,
				Version: tt.fields.Version,
			}
			agent.stop()
		})
	}
}

func TestAgent_getStorageURI(t *testing.T) {
	type fields struct {
		App     fyne.App
		Tray    fyne.Window
		Name    string
		Version string
	}
	tests := []struct {
		name   string
		fields fields
		want   fyne.URI
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				app:     tt.fields.App,
				tray:    tt.fields.Tray,
				Name:    tt.fields.Name,
				Version: tt.fields.Version,
			}
			if got := agent.getStorageURI(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Agent.getStorageURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_tracker(t *testing.T) {
	type fields struct {
		App     fyne.App
		Tray    fyne.Window
		Name    string
		Version string
	}
	type args struct {
		agentCtx context.Context
		configWG *sync.WaitGroup
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
				app:     tt.fields.App,
				tray:    tt.fields.Tray,
				Name:    tt.fields.Name,
				Version: tt.fields.Version,
			}
			agent.tracker(tt.args.agentCtx, tt.args.configWG)
		})
	}
}
