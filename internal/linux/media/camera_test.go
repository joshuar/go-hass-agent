// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package media

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladimirvivien/go4vl/device"
)

func skipCI(t *testing.T) {
	t.Helper()

	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func skipContainer(t *testing.T) {
	t.Helper()

	if os.Getenv("DEVCONTAINER") != "" {
		t.Skip("Skipping testing in dev container environment")
	}
}

func TestNewCameraControl(t *testing.T) {
	msgCh := make(chan *mqttapi.Msg)
	defer close(msgCh)
	type args struct {
		ctx          context.Context
		msgCh        chan *mqttapi.Msg
		parentLogger *slog.Logger
		mqttDevice   *mqtthass.Device
	}
	tests := []struct {
		args args
		want *CameraEntities
		name string
	}{
		{
			name: "successful",
			args: args{ctx: context.TODO(), msgCh: msgCh, parentLogger: slog.Default(), mqttDevice: &mqtthass.Device{Name: "test"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCameraControl(tt.args.ctx, tt.args.msgCh, tt.args.parentLogger, tt.args.mqttDevice)
			assert.NotNil(t, got.Status)
			state := <-msgCh
			assert.Equal(t, []byte(stoppedState), state.Message)
			assert.NotNil(t, got.Images)
			assert.NotNil(t, got.StartButton)
			assert.NotNil(t, got.StopButton)
		})
	}
}

func Test_newCamera(t *testing.T) {
	type args struct {
		logger *slog.Logger
	}
	tests := []struct {
		args args
		want *cameraControl
		name string
	}{
		{
			name: "successful",
			args: args{logger: slog.Default()},
			want: &cameraControl{
				state:  stoppedState,
				logger: slog.Default().WithGroup("camera"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newCamera(tt.args.logger)
			assert.NotNil(t, got.logger)
			assert.Equal(t, stoppedState, got.state)
		})
	}
}

func Test_cameraControl_openCamera(t *testing.T) {
	skipCI(t)
	skipContainer(t)
	type fields struct {
		device     *device.Device
		cancelFunc context.CancelFunc
		logger     *slog.Logger
		state      string
		fps        time.Duration
	}
	type args struct {
		cameraDevice string
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid device",
			args:   args{cameraDevice: defaultDevice},
			fields: fields{logger: slog.Default()},
		},
		{
			name:    "invalid device",
			args:    args{cameraDevice: "/dev/null"},
			fields:  fields{logger: slog.Default()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cameraControl{
				device:     tt.fields.device,
				cancelFunc: tt.fields.cancelFunc,
				logger:     tt.fields.logger,
				state:      tt.fields.state,
				fps:        tt.fields.fps,
			}
			if err := c.openCamera(tt.args.cameraDevice); (err != nil) != tt.wantErr {
				t.Errorf("cameraControl.openCamera() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.NotNil(t, c.device)
				assert.NotZero(t, c.fps)
				err := c.closeCamera()
				require.NoError(t, err)
			}
		})
	}
}

func Test_cameraControl_publishImages(t *testing.T) {
	skipCI(t)
	skipContainer(t)
	camera := newCamera(slog.Default())
	err := camera.openCamera(defaultDevice)
	require.NoError(t, err)
	defer camera.closeCamera() //nolint:errcheck
	msgCh := make(chan *mqttapi.Msg)
	ctx, cancelFunc := context.WithCancel(context.TODO())
	camera.cancelFunc = cancelFunc

	type args struct {
		ctx   context.Context
		msgCh chan *mqttapi.Msg
		topic string
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "successful",
			args: args{ctx: ctx, msgCh: msgCh},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				t.Log("Getting camera images...")
				var i int
				for range msgCh {
					if i == 60 {
						camera.cancelFunc()
					}
					i++
				}
			}()
			go func() {
				defer close(msgCh)
				camera.publishImages(tt.args.ctx, tt.args.topic, tt.args.msgCh) //revive:disable:datarace
			}()
			wg.Wait()
		})
	}
}

func Test_cameraControl_closeCamera(t *testing.T) {
	skipCI(t)
	skipContainer(t)
	camera := newCamera(slog.Default())
	err := camera.openCamera(defaultDevice)
	require.NoError(t, err)

	type fields struct {
		camera *cameraControl
	}
	tests := []struct {
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:   "successful",
			fields: fields{camera: camera},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.camera
			if err := c.closeCamera(); (err != nil) != tt.wantErr {
				t.Errorf("cameraControl.closeCamera() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
