// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:containedctx,dupl,exhaustruct,nlreturn,paralleltest,wsl
//revive:disable:unused-receiver
package commands

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/eclipse/paho.golang/paho"
	"github.com/go-test/deep"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"
	"github.com/stretchr/testify/require"
)

var mockCommandCallback = func(_ *paho.Publish) {}

var mockButton = mqtthass.AsButton(
	mqtthass.NewEntity("test", "test button", "test_button").
		WithOriginInfo(&mqtthass.Origin{}).
		WithDeviceInfo(&mqtthass.Device{}).
		WithIcon("mdi:test").
		WithCommandCallback(mockCommandCallback))

var mockSwitch = mqtthass.AsSwitch(
	mqtthass.NewEntity("test", "test switch", "test_switch").
		WithOriginInfo(&mqtthass.Origin{}).
		WithDeviceInfo(&mqtthass.Device{}).
		WithIcon("mdi:test").
		WithCommandCallback(mockCommandCallback), true)

func TestController_Subscriptions(t *testing.T) {
	var mockButtonSubscription, mockSwitchSubscription *mqttapi.Subscription
	var err error

	mockButtonSubscription, err = mockButton.MarshalSubscription()
	require.NoError(t, err)
	mockSwitchSubscription, err = mockSwitch.MarshalSubscription()
	require.NoError(t, err)

	type fields struct {
		buttons  []*mqtthass.ButtonEntity
		switches []*mqtthass.SwitchEntity
	}
	tests := []struct {
		name   string
		fields fields
		want   []*mqttapi.Subscription
	}{
		{
			name:   "with subscriptions",
			fields: fields{buttons: []*mqtthass.ButtonEntity{mockButton}, switches: []*mqtthass.SwitchEntity{mockSwitch}},
			want:   []*mqttapi.Subscription{mockButtonSubscription, mockSwitchSubscription},
		},
		{
			name: "without subscriptions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Controller{
				buttons:  tt.fields.buttons,
				switches: tt.fields.switches,
			}
			got := d.Subscriptions()
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestController_Configs(t *testing.T) {
	var mockButtonConfig, mockSwitchConfig *mqttapi.Msg
	var err error

	mockButtonConfig, err = mockButton.MarshalConfig()
	require.NoError(t, err)
	mockSwitchConfig, err = mockSwitch.MarshalConfig()
	require.NoError(t, err)

	type fields struct {
		buttons  []*mqtthass.ButtonEntity
		switches []*mqtthass.SwitchEntity
	}
	tests := []struct {
		name   string
		fields fields
		want   []*mqttapi.Msg
	}{
		{
			name:   "with configs",
			fields: fields{buttons: []*mqtthass.ButtonEntity{mockButton}, switches: []*mqtthass.SwitchEntity{mockSwitch}},
			want:   []*mqttapi.Msg{mockButtonConfig, mockSwitchConfig},
		},
		{
			name: "without configs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Controller{
				buttons:  tt.fields.buttons,
				switches: tt.fields.switches,
			}
			got := d.Configs()
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestNewCommandsController(t *testing.T) {
	// valid commands file
	validCommandsFile, err := filepath.Abs("testdata/commands.toml")
	require.NoError(t, err)

	// invalid commands file
	invalidCommandsFile, err := filepath.Abs("testdata/invalidcommands.toml")
	require.NoError(t, err)

	// unreadable commands file
	unreadableCommandsFile := filepath.Join(t.TempDir(), "unreadable.toml")
	_, err = exec.Command("touch", unreadableCommandsFile).Output()
	require.NoError(t, err)
	_, err = exec.Command("chmod", "a-r", unreadableCommandsFile).Output()
	require.NoError(t, err)

	mockDevice := &mqtthass.Device{}

	type args struct {
		ctx          context.Context
		device       *mqtthass.Device
		commandsFile string
	}
	tests := []struct {
		want    *Controller
		args    args
		name    string
		wantErr bool
	}{
		{
			name:    "no commands file",
			wantErr: true,
		},
		{
			name:    "unreadable commands file",
			wantErr: true,
			args:    args{ctx: context.TODO(), commandsFile: unreadableCommandsFile, device: mockDevice},
		},
		{
			name:    "invalid commands file",
			wantErr: true,
			args:    args{ctx: context.TODO(), commandsFile: invalidCommandsFile, device: mockDevice},
		},
		{
			name: "valid commands file",
			args: args{ctx: context.TODO(), commandsFile: validCommandsFile, device: mockDevice},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCommandsController(tt.args.ctx, tt.args.commandsFile, tt.args.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCommandsController() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_buttonCmd(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successful command",
			args: args{command: "true"},
		},
		{
			name:    "unsuccessful command",
			args:    args{command: "false"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := buttonCmd(tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("buttonCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_switchCmd(t *testing.T) {
	type args struct {
		command string
		state   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successful command",
			args: args{command: "true", state: "someState"},
		},
		{
			name:    "unsuccessful command",
			args:    args{command: "false", state: "someState"},
			wantErr: true,
		},
		{
			name:    "no state",
			args:    args{command: "true"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := switchCmd(tt.args.command, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("switchCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_switchState(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name: "switch ON",
			args: args{command: "echo ON"},
			want: json.RawMessage(`ON`),
		},
		{
			name: "switch OFF",
			args: args{command: "echo OFF"},
			want: json.RawMessage(`OFF`),
		},
		{
			name:    "unsuccessful",
			args:    args{command: "false"},
			wantErr: true,
		},
		{
			name:    "unknown output",
			args:    args{command: "echo SOMETHING"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := switchState(tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("switchState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("switchState() = %v, want %v", got, tt.want)
			}
		})
	}
}
