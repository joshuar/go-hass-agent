// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/api"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/stretchr/testify/assert"
)

func TestSensorTracker_add(t *testing.T) {
	mockSensor := &SensorMock{
		IDFunc: func() string { return "sensorID" },
	}

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.RWMutex
	}
	type args struct {
		s Sensor
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful add",
			fields: fields{
				sensor: make(map[string]Sensor),
			},
			args:    args{s: mockSensor},
			wantErr: false,
		},
		{
			name:    "unsuccessful add (not initialised properly)",
			args:    args{s: mockSensor},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			if err := tracker.add(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSensorTracker_Get(t *testing.T) {
	mockSensor := &SensorMock{}
	mockMap := make(map[string]Sensor)
	mockMap["sensorID"] = mockSensor

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.RWMutex
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Sensor
		wantErr bool
	}{
		{
			name:    "successful get",
			fields:  fields{sensor: mockMap},
			args:    args{id: "sensorID"},
			wantErr: false,
		},
		{
			name:    "unsuccessful get",
			fields:  fields{sensor: mockMap},
			args:    args{id: "doesntExist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			_, err := tracker.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("SensorTracker.Get() = %v, want %v", got, tt.want)
			// }
		})
	}
}

type mockConfig struct {
	url     string
	storage string
}

func (c *mockConfig) WebSocketURL() string {
	return ""
}

func (c *mockConfig) WebhookID() string {
	return ""
}
func (c *mockConfig) Token() string {
	return ""
}
func (c *mockConfig) ApiURL() string {
	return c.url
}
func (c *mockConfig) Secret() string {
	return ""
}

func (c *mockConfig) NewStorage(id string) (string, error) {
	return c.storage, nil
}

func NewMockConfig(t *testing.T) *mockConfig {
	path, err := os.MkdirTemp("/tmp", "go-hass-agent-test")
	assert.Nil(t, err)
	return &mockConfig{
		storage: path,
	}
}

func TestSensorTracker_Update(t *testing.T) {
	mockServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &api.UnencryptedRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.Nil(t, err)
			switch req.Type {
			case "update_sensor_states":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
			}
		}))
	}
	server := mockServer(t)
	// defer server.Close()

	mockRegistry := &RegistryMock{
		IsRegisteredFunc: func(s string) chan bool {
			valueCh := make(chan bool, 1)
			valueCh <- true
			return valueCh
		},
		IsDisabledFunc: func(s string) chan bool {
			valueCh := make(chan bool, 1)
			valueCh <- false
			return valueCh
		},
	}

	mockSensorUpdate := &SensorMock{
		AttributesFunc: func() interface{} { return nil },
		StateFunc:      func() interface{} { return "aState" },
		IconFunc:       func() string { return "mdi:icon" },
		SensorTypeFunc: func() sensor.SensorType { return sensor.TypeSensor },
		IDFunc:         func() string { return "sensorID" },
		NameFunc:       func() string { return "sensorName" },
	}

	mockConfig := &mockConfig{
		url: server.URL,
	}

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.RWMutex
	}
	type args struct {
		ctx          context.Context
		config       api.Config
		sensorUpdate Sensor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "successful update",
			fields: fields{
				sensor:   make(map[string]Sensor),
				registry: mockRegistry,
			},
			args: args{
				ctx:          context.Background(),
				config:       mockConfig,
				sensorUpdate: mockSensorUpdate,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			tracker.updateSensor(tt.args.ctx, tt.args.config, tt.args.sensorUpdate)
		})
	}
}

// func TestRunSensorTracker(t *testing.T) {
// 	mockConfig := NewMockConfig(t)
// 	defer os.RemoveAll(mockConfig.storage)

// 	ctx, cancelfunc := context.WithCancel(context.Background())
// 	defer cancelfunc()

// 	type args struct {
// 		ctx    context.Context
// 		config api.Config
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "successful test",
// 			args: args{
// 				ctx:    ctx,
// 				config: mockConfig,
// 			},
// 			wantErr: false,
// 		},
// 		// 		// {
// 		// 		// 	name: "successful test, no directory given",
// 		// 		// 	args: args{
// 		// 		// 		ctx:  context.Background(),
// 		// 		// 		path: "",
// 		// 		// 	},
// 		// 		// 	wantErr: false,
// 		// 		// },
// 		// 		// {
// 		// 		// 	name: "unsuccessful test, invalid directory",
// 		// 		// 	args: args{
// 		// 		// 		ctx:  context.Background(),
// 		// 		// 		path: "/foo/bar",
// 		// 		// 	},
// 		// 		// 	wantErr: true,
// 		// 		// },
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := RunSensorTracker(tt.args.ctx, tt.args.config); (err != nil) != tt.wantErr {
// 				t.Errorf("RunSensorTracker() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
