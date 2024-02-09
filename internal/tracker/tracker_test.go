// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
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
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var defaultTestPrefs = []preferences.Preference{
	preferences.Token("testToken"),
	preferences.CloudhookURL(""),
	preferences.RemoteUIURL(""),
	preferences.WebhookID("testID"),
	preferences.Secret(""),
	preferences.DeviceName("testDevice"),
	preferences.DeviceID("testID"),
	preferences.Version("6.4.0"),
	preferences.Registered(true),
}

func mockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		raw := struct {
			Type string `json:"type"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&raw)
		assert.Nil(t, err)
		switch raw.Type {
		case "update_sensor_states":
			upd := &sensor.SensorUpdateInfo{}
			assert.Nil(t, err)
			json.NewDecoder(r.Body).Decode(&upd)
			assert.Nil(t, err)
			resp := "{" + `"` + upd.UniqueID + `"` + `:{"success":true}}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(resp))
		case "register_sensor":
			reg := &sensor.SensorRegistrationInfo{}
			json.NewDecoder(r.Body).Decode(&reg)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}
	}))
}

func TestSensorTracker_add(t *testing.T) {
	mockSensor := &SensorMock{
		IDFunc: func() string { return "sensorID" },
	}

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.Mutex
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
			tr := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			if err := tr.add(tt.args.s); (err != nil) != tt.wantErr {
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
		mu       sync.Mutex
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
			want:    mockSensor,
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
			tr := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			got, err := tr.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorTracker.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorTracker_SensorList(t *testing.T) {
	mockSensor := &SensorMock{
		StateFunc: func() any { return "aState" },
	}
	mockMap := make(map[string]Sensor)
	mockMap["sensorID"] = mockSensor

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "with sensors",
			fields: fields{sensor: mockMap},
			want:   []string{"sensorID"},
		},
		{
			name: "without sensors",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			if got := tr.SensorList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorTracker.SensorList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorTracker_send(t *testing.T) {
	mockServer := mockServer(t)
	defer mockServer.Close()

	preferences.SetPath(t.TempDir())
	prefs := defaultTestPrefs
	prefs = append(prefs,
		preferences.Host(mockServer.URL),
		preferences.RestAPIURL(mockServer.URL),
		preferences.WebsocketURL(mockServer.URL),
	)
	err := preferences.Save(prefs...)
	assert.Nil(t, err)
	p, err := preferences.Load()
	assert.Nil(t, err)
	ctx := preferences.EmbedInContext(context.TODO(), p)

	mockUpdate := &SensorMock{
		IDFunc:         func() string { return "updateID" },
		NameFunc:       func() string { return "Update Sensor" },
		UnitsFunc:      func() string { return "" },
		StateFunc:      func() any { return "aState" },
		AttributesFunc: func() any { return nil },
		IconFunc:       func() string { return "anIcon" },
		SensorTypeFunc: func() sensor.SensorType { return sensor.TypeSensor },
	}
	mockRegistration := &SensorMock{
		IDFunc:          func() string { return "regID" },
		NameFunc:        func() string { return "Registration Sensor" },
		UnitsFunc:       func() string { return "" },
		StateFunc:       func() any { return "aState" },
		AttributesFunc:  func() any { return nil },
		IconFunc:        func() string { return "anIcon" },
		SensorTypeFunc:  func() sensor.SensorType { return sensor.TypeSensor },
		DeviceClassFunc: func() sensor.SensorDeviceClass { return sensor.Duration },
		StateClassFunc:  func() sensor.SensorStateClass { return sensor.StateMeasurement },
		CategoryFunc:    func() string { return "" },
	}
	mockDisabled := &SensorMock{
		IDFunc:          func() string { return "disabledID" },
		NameFunc:        func() string { return "Update Sensor" },
		UnitsFunc:       func() string { return "" },
		StateFunc:       func() any { return "aState" },
		AttributesFunc:  func() any { return nil },
		IconFunc:        func() string { return "anIcon" },
		SensorTypeFunc:  func() sensor.SensorType { return sensor.TypeSensor },
		DeviceClassFunc: func() sensor.SensorDeviceClass { return sensor.Duration },
		StateClassFunc:  func() sensor.SensorStateClass { return sensor.StateMeasurement },
		CategoryFunc:    func() string { return "" },
	}
	mockMap := make(map[string]Sensor)
	mockMap["updateID"] = mockUpdate
	mockMap["regID"] = mockRegistration
	mockMap["disabledID"] = mockDisabled
	mockRegistry := &RegistryMock{
		IsDisabledFunc: func(s string) chan bool {
			d := make(chan bool, 1)
			switch s {
			case "disabledID":
				d <- true
			default:
				d <- false
			}
			close(d)
			return d
		},
		IsRegisteredFunc: func(s string) chan bool {
			d := make(chan bool, 1)
			switch s {
			case "updateID":
				d <- true
			case "regID":
				d <- false
			}
			close(d)
			return d
		},
		SetRegisteredFunc: func(s string, b bool) error {
			return nil
		},
	}

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.Mutex
	}
	type args struct {
		ctx          context.Context
		sensorUpdate Sensor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "sensor update",
			fields: fields{sensor: mockMap, registry: mockRegistry},
			args:   args{ctx: ctx, sensorUpdate: mockUpdate},
		},
		{
			name:   "sensor registration",
			fields: fields{sensor: mockMap, registry: mockRegistry},
			args:   args{ctx: ctx, sensorUpdate: mockRegistration},
		},
		{
			name:   "disabled sensor",
			fields: fields{sensor: mockMap, registry: mockRegistry},
			args:   args{ctx: ctx, sensorUpdate: mockDisabled},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			tr.send(tt.args.ctx, tt.args.sensorUpdate)
		})
	}
}

func TestSensorTracker_handle(t *testing.T) {
	mockUpdate := &SensorMock{
		IDFunc:         func() string { return "updateID" },
		NameFunc:       func() string { return "Update Sensor" },
		UnitsFunc:      func() string { return "" },
		StateFunc:      func() any { return "aState" },
		AttributesFunc: func() any { return nil },
		IconFunc:       func() string { return "anIcon" },
		SensorTypeFunc: func() sensor.SensorType { return sensor.TypeSensor },
	}
	mockUpdateResponse := &apiResponseMock{
		TypeFunc:     func() api.ResponseType { return api.ResponseTypeUpdate },
		DisabledFunc: func() bool { return false },
	}
	mockRegistration := &SensorMock{
		IDFunc:          func() string { return "regID" },
		NameFunc:        func() string { return "Registration Sensor" },
		UnitsFunc:       func() string { return "" },
		StateFunc:       func() any { return "aState" },
		AttributesFunc:  func() any { return nil },
		IconFunc:        func() string { return "anIcon" },
		SensorTypeFunc:  func() sensor.SensorType { return sensor.TypeSensor },
		DeviceClassFunc: func() sensor.SensorDeviceClass { return sensor.Duration },
		StateClassFunc:  func() sensor.SensorStateClass { return sensor.StateMeasurement },
		CategoryFunc:    func() string { return "" },
	}
	mockRegistrationResponse := &apiResponseMock{
		TypeFunc:       func() api.ResponseType { return api.ResponseTypeRegistration },
		RegisteredFunc: func() bool { return true },
	}
	mockDisabled := &SensorMock{
		IDFunc:         func() string { return "disabledID" },
		NameFunc:       func() string { return "Disabled Sensor" },
		UnitsFunc:      func() string { return "" },
		StateFunc:      func() any { return "aState" },
		AttributesFunc: func() any { return nil },
		IconFunc:       func() string { return "anIcon" },
		SensorTypeFunc: func() sensor.SensorType { return sensor.TypeSensor },
	}
	mockDisabledResponse := &apiResponseMock{
		TypeFunc:     func() api.ResponseType { return api.ResponseTypeUpdate },
		DisabledFunc: func() bool { return true },
	}
	mockMap := make(map[string]Sensor)
	mockMap["updateID"] = mockUpdate
	mockMap["regID"] = mockRegistration
	mockMap["disabledID"] = mockDisabled
	mockRegistry := &RegistryMock{
		IsDisabledFunc: func(s string) chan bool {
			d := make(chan bool, 1)
			d <- false
			close(d)
			return d
		},
		IsRegisteredFunc: func(s string) chan bool {
			d := make(chan bool, 1)
			switch s {
			case "updateID":
				d <- true
			case "regID":
				d <- false
			}
			close(d)
			return d
		},
		SetRegisteredFunc: func(s string, b bool) error {
			return nil
		},
		SetDisabledFunc: func(s string, b bool) error {
			return nil
		},
	}

	type fields struct {
		registry Registry
		sensor   map[string]Sensor
		mu       sync.Mutex
	}
	type args struct {
		response     apiResponse
		sensorUpdate Sensor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "successful update",
			args:   args{response: mockUpdateResponse, sensorUpdate: mockUpdate},
			fields: fields{sensor: mockMap, registry: mockRegistry},
		},
		{
			name:   "successful registration",
			args:   args{response: mockRegistrationResponse, sensorUpdate: mockRegistration},
			fields: fields{sensor: mockMap, registry: mockRegistry},
		},
		{
			name:   "disabled update",
			args:   args{response: mockDisabledResponse, sensorUpdate: mockDisabled},
			fields: fields{sensor: mockMap, registry: mockRegistry},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &SensorTracker{
				registry: tt.fields.registry,
				sensor:   tt.fields.sensor,
				mu:       tt.fields.mu,
			}
			tr.handle(tt.args.response, tt.args.sensorUpdate)
		})
	}
}

func TestNewSensorTracker(t *testing.T) {
	testID := "go-hass-agent-test"
	basePath = t.TempDir()
	assert.Nil(t, os.Mkdir(filepath.Join(basePath, testID), 0o755))
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *SensorTracker
		wantErr bool
	}{
		{
			name: "default test",
			args: args{id: testID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSensorTracker(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSensorTracker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewSensorTracker() = %v, want %v", got, tt.want)
			// }
		})
	}
}
