// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
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
	"github.com/stretchr/testify/require"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"
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

var mockNumber = mqtthass.AsNumber(
	mqtthass.NewEntity("test", "test number", "test_number").
		WithOriginInfo(&mqtthass.Origin{}).
		WithDeviceInfo(&mqtthass.Device{}).
		WithIcon("mdi:test").
		WithCommandCallback(mockCommandCallback), 0, 100, 1, mqtthass.NumberAuto)

func TestController_Subscriptions(t *testing.T) {
	var mockButtonSubscription, mockSwitchSubscription, mockNumberSubscription *mqttapi.Subscription
	var err error

	mockButtonSubscription, err = mockButton.MarshalSubscription()
	require.NoError(t, err)
	mockSwitchSubscription, err = mockSwitch.MarshalSubscription()
	require.NoError(t, err)
	mockNumberSubscription, err = mockNumber.MarshalSubscription()
	require.NoError(t, err)

	type fields struct {
		buttons    []*mqtthass.ButtonEntity
		switches   []*mqtthass.SwitchEntity
		intNumbers []*mqtthass.NumberEntity[int]
	}
	tests := []struct {
		name   string
		fields fields
		want   []*mqttapi.Subscription
	}{
		{
			name: "with subscriptions",
			fields: fields{
				buttons:    []*mqtthass.ButtonEntity{mockButton},
				switches:   []*mqtthass.SwitchEntity{mockSwitch},
				intNumbers: []*mqtthass.NumberEntity[int]{mockNumber},
			},
			want: []*mqttapi.Subscription{
				mockButtonSubscription,
				mockSwitchSubscription,
				mockNumberSubscription,
			},
		},
		{
			name: "without subscriptions",
			want: []*mqttapi.Subscription{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Controller{
				buttons:    tt.fields.buttons,
				switches:   tt.fields.switches,
				intNumbers: tt.fields.intNumbers,
			}
			got := d.Subscriptions()
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestController_Configs(t *testing.T) {
	var mockButtonConfig, mockSwitchConfig, mockNumberConfig *mqttapi.Msg
	var err error

	mockButtonConfig, err = mockButton.MarshalConfig()
	require.NoError(t, err)
	mockSwitchConfig, err = mockSwitch.MarshalConfig()
	require.NoError(t, err)
	mockNumberConfig, err = mockNumber.MarshalConfig()
	require.NoError(t, err)

	type fields struct {
		buttons    []*mqtthass.ButtonEntity
		switches   []*mqtthass.SwitchEntity
		intNumbers []*mqtthass.NumberEntity[int]
	}
	tests := []struct {
		name   string
		fields fields
		want   []*mqttapi.Msg
	}{
		{
			name: "with configs",
			fields: fields{
				buttons:    []*mqtthass.ButtonEntity{mockButton},
				switches:   []*mqtthass.SwitchEntity{mockSwitch},
				intNumbers: []*mqtthass.NumberEntity[int]{mockNumber},
			},
			want: []*mqttapi.Msg{
				mockButtonConfig,
				mockSwitchConfig,
				mockNumberConfig,
			},
		},
		{
			name: "without configs",
			want: []*mqttapi.Msg{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Controller{
				buttons:    tt.fields.buttons,
				switches:   tt.fields.switches,
				intNumbers: tt.fields.intNumbers,
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
			if err := cmdWithState(tt.args.command, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("switchCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cmdWithoutState(t *testing.T) {
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
			if err := cmdWithoutState(tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("cmdWithoutState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cmdWithState(t *testing.T) {
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
			args: args{command: "true", state: "foo"},
		},
		{
			name:    "unsuccessful command",
			args:    args{command: "false", state: "foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cmdWithState(tt.args.command, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("cmdWithState() error = %v, wantErr %v", err, tt.wantErr)
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

func Test_numberState(t *testing.T) {
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
			name: "valid int",
			args: args{command: "echo 1"},
			want: json.RawMessage(`{ "value": 1 }`),
		},
		{
			name: "valid float",
			args: args{command: "echo 0.567"},
			want: json.RawMessage(`{ "value": 0.567 }`),
		},
		{
			name:    "invalid (quoted)",
			args:    args{command: `echo "1"`},
			wantErr: true,
		},
		{
			name:    "invalid (string)",
			args:    args{command: `echo some string`},
			wantErr: true,
		},
		{
			name:    "invalid (multiple outputs)",
			args:    args{command: `echo 0.6788 1.3453`},
			wantErr: true,
		},
		{
			name:    "unsuccessful",
			args:    args{command: "false"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := numberState(tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("numberState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("numberState() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
