// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package sensor

import (
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

func TestState_UpdateValue(t *testing.T) {
	type fields struct {
		Value      any
		Attributes map[string]any
		Icon       string
	}
	type args struct {
		value any
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
			s := &State{
				Value:      tt.fields.Value,
				Attributes: tt.fields.Attributes,
				Icon:       tt.fields.Icon,
			}
			s.UpdateValue(tt.args.value)
		})
	}
}

func TestState_UpdateIcon(t *testing.T) {
	type fields struct {
		Value      any
		Attributes map[string]any
		Icon       string
	}
	type args struct {
		icon string
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
			s := &State{
				Value:      tt.fields.Value,
				Attributes: tt.fields.Attributes,
				Icon:       tt.fields.Icon,
			}
			s.UpdateIcon(tt.args.icon)
		})
	}
}

func TestState_UpdateAttribute(t *testing.T) {
	type fields struct {
		Value      any
		Attributes map[string]any
		Icon       string
	}
	type args struct {
		key   string
		value any
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
			s := &State{
				Value:      tt.fields.Value,
				Attributes: tt.fields.Attributes,
				Icon:       tt.fields.Icon,
			}
			s.UpdateAttribute(tt.args.key, tt.args.value)
		})
	}
}

func TestState_Validate(t *testing.T) {
	type fields struct {
		Value      any
		Attributes map[string]any
		Icon       string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				Value:      tt.fields.Value,
				Attributes: tt.fields.Attributes,
				Icon:       tt.fields.Icon,
			}
			if err := s.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("State.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSensor(t *testing.T) {
	type args struct {
		options []Option[Entity]
	}
	tests := []struct {
		name string
		args args
		want Entity
	}{
		{
			name: "sensor",
			args: args{
				[]Option[Entity]{
					WithName("Mock Entity"),
					WithID("mock_entity"),
					AsTypeSensor(),
					WithUnits("units"),
					AsDiagnostic(),
					WithState(
						WithValue("value"),
						WithIcon("mdi:icon"),
					),
				},
			},
			want: Entity{
				Name:       "Mock Entity",
				ID:         "mock_entity",
				EntityType: types.Sensor,
				Units:      "units",
				Category:   types.CategoryDiagnostic,
				State: &State{
					Value: "value",
					Icon:  "mdi:icon",
				},
			},
		},
		{
			name: "binary_sensor",
			args: args{
				[]Option[Entity]{
					WithName("Mock Entity"),
					WithID("mock_entity"),
					AsTypeBinarySensor(),
					WithUnits("units"),
					AsDiagnostic(),
					WithState(
						WithValue("value"),
						WithIcon("mdi:icon"),
					),
				},
			},
			want: Entity{
				Name:       "Mock Entity",
				ID:         "mock_entity",
				EntityType: types.BinarySensor,
				Units:      "units",
				Category:   types.CategoryDiagnostic,
				State: &State{
					Value: "value",
					Icon:  "mdi:icon",
				},
			},
		},
		{
			name: "set_single_attribute",
			args: args{
				[]Option[Entity]{
					WithName("Mock Entity"),
					WithID("mock_entity"),
					WithState(
						WithValue("value"),
						WithIcon("mdi:icon"),
						WithAttribute("attr", "attr_value"),
					),
				},
			},
			want: Entity{
				Name: "Mock Entity",
				ID:   "mock_entity",
				State: &State{
					Value:      "value",
					Icon:       "mdi:icon",
					Attributes: map[string]any{"attr": "attr_value"},
				},
			},
		},
		{
			name: "set_multiple_attributes",
			args: args{
				[]Option[Entity]{
					WithName("Mock Entity"),
					WithID("mock_entity"),
					WithState(
						WithValue("value"),
						WithIcon("mdi:icon"),
						WithAttributes(map[string]any{"attr1": "attr_value", "attr2": "attr_value"}),
					),
				},
			},
			want: Entity{
				Name: "Mock Entity",
				ID:   "mock_entity",
				State: &State{
					Value:      "value",
					Icon:       "mdi:icon",
					Attributes: map[string]any{"attr1": "attr_value", "attr2": "attr_value"},
				},
			},
		},
		{
			name: "set_data_source",
			args: args{
				[]Option[Entity]{
					WithName("Mock Entity"),
					WithID("mock_entity"),
					WithState(
						WithValue("value"),
						WithIcon("mdi:icon"),
						WithDataSourceAttribute("source"),
						WithAttributes(map[string]any{"attr1": "attr_value", "attr2": "attr_value"}),
					),
				},
			},
			want: Entity{
				Name: "Mock Entity",
				ID:   "mock_entity",
				State: &State{
					Value:      "value",
					Icon:       "mdi:icon",
					Attributes: map[string]any{"data_source": "source", "attr1": "attr_value", "attr2": "attr_value"},
				},
			},
		},
		{
			name: "retryable",
			args: args{
				[]Option[Entity]{
					WithName("Mock Entity"),
					WithID("mock_entity"),
					AsTypeSensor(),
					WithUnits("units"),
					AsDiagnostic(),
					WithState(
						WithValue("value"),
						WithIcon("mdi:icon"),
					),
					WithRequestRetry(true),
				},
			},
			want: Entity{
				Name:       "Mock Entity",
				ID:         "mock_entity",
				EntityType: types.Sensor,
				Units:      "units",
				Category:   types.CategoryDiagnostic,
				State: &State{
					Value: "value",
					Icon:  "mdi:icon",
				},
				requestMetadata: requestMetadata{RetryRequest: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSensor(tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSensor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntity_UpdateState(t *testing.T) {
	type fields struct {
		State           *State
		requestMetadata requestMetadata
		ID              string
		Name            string
		Units           string
		EntityType      types.SensorType
		DeviceClass     types.DeviceClass
		StateClass      types.StateClass
		Category        types.Category
	}
	type args struct {
		options []Option[State]
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
			e := &Entity{
				State:           tt.fields.State,
				requestMetadata: tt.fields.requestMetadata,
				ID:              tt.fields.ID,
				Name:            tt.fields.Name,
				Units:           tt.fields.Units,
				EntityType:      tt.fields.EntityType,
				DeviceClass:     tt.fields.DeviceClass,
				StateClass:      tt.fields.StateClass,
				Category:        tt.fields.Category,
			}
			e.UpdateState(tt.args.options...)
		})
	}
}

func TestEntity_Validate(t *testing.T) {
	type fields struct {
		State           *State
		requestMetadata requestMetadata
		ID              string
		Name            string
		Units           string
		EntityType      types.SensorType
		DeviceClass     types.DeviceClass
		StateClass      types.StateClass
		Category        types.Category
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entity{
				State:           tt.fields.State,
				requestMetadata: tt.fields.requestMetadata,
				ID:              tt.fields.ID,
				Name:            tt.fields.Name,
				Units:           tt.fields.Units,
				EntityType:      tt.fields.EntityType,
				DeviceClass:     tt.fields.DeviceClass,
				StateClass:      tt.fields.StateClass,
				Category:        tt.fields.Category,
			}
			if err := e.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Entity.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocation_Validate(t *testing.T) {
	type fields struct {
		Gps              []float64
		GpsAccuracy      int
		Battery          int
		Speed            int
		Altitude         int
		Course           int
		VerticalAccuracy int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Location{
				Gps:              tt.fields.Gps,
				GpsAccuracy:      tt.fields.GpsAccuracy,
				Battery:          tt.fields.Battery,
				Speed:            tt.fields.Speed,
				Altitude:         tt.fields.Altitude,
				Course:           tt.fields.Course,
				VerticalAccuracy: tt.fields.VerticalAccuracy,
			}
			if err := l.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Location.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
