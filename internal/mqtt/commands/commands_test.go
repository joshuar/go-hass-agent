// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package commands

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/adrg/xdg"
	"github.com/eclipse/paho.golang/paho"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/require"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

var mockCommandCallback = func(_ *paho.Publish) {}

var mockButton = mqtthass.NewButtonEntity().
	WithDetails(
		mqtthass.App("test"),
		mqtthass.Name("test button"),
		mqtthass.ID("test_button"),
		mqtthass.OriginInfo(&mqtthass.Origin{}),
		mqtthass.DeviceInfo(&mqtthass.Device{}),
		mqtthass.Icon("mdi:test"),
	).
	WithCommand(
		mqtthass.CommandCallback(mockCommandCallback),
	)

var mockSwitch = mqtthass.NewSwitchEntity().
	WithDetails(
		mqtthass.App("test"),
		mqtthass.Name("test switch"),
		mqtthass.ID("test_switch"),
		mqtthass.OriginInfo(&mqtthass.Origin{}),
		mqtthass.DeviceInfo(&mqtthass.Device{}),
		mqtthass.Icon("mdi:test"),
	).
	WithCommand(
		mqtthass.CommandCallback(mockCommandCallback),
	).
	OptimisticMode()

var mockNumber = mqtthass.NewNumberEntity[int64]().
	WithDetails(
		mqtthass.App("test"),
		mqtthass.Name("test number"),
		mqtthass.ID("test_number"),
		mqtthass.OriginInfo(&mqtthass.Origin{}),
		mqtthass.DeviceInfo(&mqtthass.Device{}),
		mqtthass.Icon("mdi:test"),
	).
	WithCommand(
		mqtthass.CommandCallback(mockCommandCallback),
	).
	WithStep(1).
	WithMin(0).
	WithMax(100).
	WithMode(mqtthass.NumberAuto)

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
		intNumbers []*mqtthass.NumberEntity[int64]
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
				intNumbers: []*mqtthass.NumberEntity[int64]{mockNumber},
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
			d := &Worker{
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
		intNumbers []*mqtthass.NumberEntity[int64]
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
				intNumbers: []*mqtthass.NumberEntity[int64]{mockNumber},
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
			d := &Worker{
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
	mockDevice := &mqtthass.Device{}

	type args struct {
		device    *mqtthass.Device
		configDir string
	}
	tests := []struct {
		want    *Worker
		args    args
		name    string
		wantErr bool
	}{
		{
			name:    "unreadable commands file",
			wantErr: true,
			args:    args{configDir: "testdata/unreadable", device: mockDevice},
		},
		{
			name:    "invalid commands file",
			wantErr: true,
			args:    args{configDir: "testdata/invalid", device: mockDevice},
		},
		{
			name: "valid commands file",
			args: args{configDir: "testdata/valid", device: mockDevice},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := preferences.PathToCtx(t.Context(), tt.args.configDir)
			xdg.ConfigHome = tt.args.configDir
			_, err := NewCommandsWorker(ctx, tt.args.device)
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
