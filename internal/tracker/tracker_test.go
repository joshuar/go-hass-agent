// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
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

// func NewMockConfig(t *testing.T) *mockConfig {
// 	path, err := os.MkdirTemp("/tmp", "go-hass-agent-test")
// 	assert.Nil(t, err)
// 	return &mockConfig{
// 		storage: path,
// 	}
// }

func TestSensorTracker_Update(t *testing.T) {
	mockServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &api.UnencryptedRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.Nil(t, err)
			switch req.Type {
			case "update_sensor_states":
				switch {
				case strings.Contains(string(req.Data), "bad"):
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"sensorID":{"success":false,"error":{"code":"invalid_format","message": "Unexpected value for type"}}}`))
				case strings.Contains(string(req.Data), "disabled"):
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"sensorID":{"success":true,"is_disabled":true}}`))
				default:
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"sensorID":{"success":true}}`))
				}
			case "register_sensor":
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"success":true}`))
			}
		}))
	}
	server := mockServer(t)
	defer server.Close()

	mockExistingRegistry := &RegistryMock{
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
		SetDisabledFunc: func(s string, b bool) error {
			return nil
		},
	}

	mockNewRegistry := &RegistryMock{
		IsRegisteredFunc: func(s string) chan bool {
			valueCh := make(chan bool, 1)
			valueCh <- false
			return valueCh
		},
		IsDisabledFunc: func(s string) chan bool {
			valueCh := make(chan bool, 1)
			valueCh <- false
			return valueCh
		},
		SetDisabledFunc:   func(s string, b bool) error { return nil },
		SetRegisteredFunc: func(s string, b bool) error { return nil },
	}

	mockSensorUpdate := &SensorMock{
		AttributesFunc:  func() interface{} { return nil },
		StateFunc:       func() interface{} { return "goodState" },
		IconFunc:        func() string { return "mdi:icon" },
		SensorTypeFunc:  func() sensor.SensorType { return sensor.TypeSensor },
		IDFunc:          func() string { return "sensorID" },
		NameFunc:        func() string { return "sensorName" },
		UnitsFunc:       func() string { return "units" },
		DeviceClassFunc: func() sensor.SensorDeviceClass { return sensor.Duration },
		StateClassFunc:  func() sensor.SensorStateClass { return sensor.StateMeasurement },
		CategoryFunc:    func() string { return "" },
	}

	mockBadSensorUpdate := &SensorMock{
		AttributesFunc:  func() interface{} { return nil },
		StateFunc:       func() interface{} { return "badState" },
		IconFunc:        func() string { return "mdi:icon" },
		SensorTypeFunc:  func() sensor.SensorType { return sensor.TypeSensor },
		IDFunc:          func() string { return "sensorID" },
		NameFunc:        func() string { return "sensorName" },
		UnitsFunc:       func() string { return "units" },
		DeviceClassFunc: func() sensor.SensorDeviceClass { return sensor.Duration },
		StateClassFunc:  func() sensor.SensorStateClass { return sensor.StateMeasurement },
		CategoryFunc:    func() string { return "" },
	}

	mockDisabledSensorUpdate := &SensorMock{
		AttributesFunc:  func() interface{} { return nil },
		StateFunc:       func() interface{} { return "disabled" },
		IconFunc:        func() string { return "mdi:icon" },
		SensorTypeFunc:  func() sensor.SensorType { return sensor.TypeSensor },
		IDFunc:          func() string { return "sensorID" },
		NameFunc:        func() string { return "sensorName" },
		UnitsFunc:       func() string { return "units" },
		DeviceClassFunc: func() sensor.SensorDeviceClass { return sensor.Duration },
		StateClassFunc:  func() sensor.SensorStateClass { return sensor.StateMeasurement },
		CategoryFunc:    func() string { return "" },
	}

	mockConfig := &agentMock{
		GetConfigFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefAPIURL:
				*v = server.URL
				return nil
			case config.PrefSecret:
				*v = ""
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.RWMutex
	}
	type args struct {
		ctx          context.Context
		config       agent
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
				registry: mockExistingRegistry,
			},
			args: args{
				ctx:          context.Background(),
				config:       mockConfig,
				sensorUpdate: mockSensorUpdate,
			},
		},
		{
			name: "bad update",
			fields: fields{
				sensor:   make(map[string]Sensor),
				registry: mockExistingRegistry,
			},
			args: args{
				ctx:          context.Background(),
				config:       mockConfig,
				sensorUpdate: mockBadSensorUpdate,
			},
		},
		{
			name: "disabled update",
			fields: fields{
				sensor:   make(map[string]Sensor),
				registry: mockExistingRegistry,
			},
			args: args{
				ctx:          context.Background(),
				config:       mockConfig,
				sensorUpdate: mockDisabledSensorUpdate,
			},
		},
		{
			name: "successful new",
			fields: fields{
				sensor:   make(map[string]Sensor),
				registry: mockNewRegistry,
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
			tracker.send(tt.args.ctx, tt.args.config, tt.args.sensorUpdate)
		})
	}
}

// func Test_startWorkers(t *testing.T) {
// 	ctx, cancelFunc := context.WithCancel(context.Background())
// 	updateCh := make(chan interface{})
// 	defer close(updateCh)
// 	defer cancelFunc()
// 	mockWorker := func(context.Context, chan interface{}) {
// 		t.Log("worker ran")
// 	}
// 	w := []func(context.Context, chan interface{}){mockWorker}

// 	type args struct {
// 		ctx      context.Context
// 		workers  []func(context.Context, chan interface{})
// 		updateCh chan interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		{
// 			name: "default test",
// 			args: args{
// 				ctx:      ctx,
// 				workers:  w,
// 				updateCh: updateCh,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			startWorkers(tt.args.ctx, tt.args.workers, tt.args.updateCh)
// 		})
// 	}
// }

// func TestSensorTracker_trackUpdates(t *testing.T) {
// 	ctx, cancelFunc := context.WithCancel(context.Background())
// 	updateCh := make(chan interface{})
// 	defer close(updateCh)
// 	defer cancelFunc()

// 	type fields struct {
// 		registry Registry
// 		sensor   map[string]Sensor
// 		mu       sync.RWMutex
// 	}
// 	type args struct {
// 		ctx      context.Context
// 		config   agent
// 		updateCh chan interface{}
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		{
// 			name: "default test",
// 			args: args{
// 				ctx:      ctx,
// 				config:   &agentMock{},
// 				updateCh: updateCh,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tr := &SensorTracker{
// 				registry: tt.fields.registry,
// 				sensor:   tt.fields.sensor,
// 				mu:       tt.fields.mu,
// 			}
// 			go tr.trackUpdates(tt.args.ctx, tt.args.config, tt.args.updateCh)
// 			cancelFunc()
// 		})
// 	}
// }
