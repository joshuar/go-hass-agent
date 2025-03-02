// Package models provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package models

import (
	"encoding/json"

	"github.com/oapi-codegen/runtime"
)

// Defines values for EntityCategory.
const (
	Diagnostic EntityCategory = "diagnostic"
)

// Defines values for SensorType.
const (
	SensorTypeBinarySensor SensorType = "binary_sensor"
	SensorTypeSensor       SensorType = "sensor"
)

// Attributes defines additional custom attributes of a entity.
type Attributes map[string]interface{}

// Entity is any valid Home Assistant Entity type.
type Entity struct {
	union json.RawMessage
}

// EntityCategory is the entity category of the entity.
type EntityCategory string

// Event defines model for Event.
type Event struct {
	// Data is data of the event to fire
	Data map[string]interface{} `json:"event_data" validate:"required"`

	// Type is the type of the event to fire.
	Type string `json:"event_type" validate:"required"`

	// Retryable indicates whether requests should be retried when sending this event data to Home Assistant.
	Retryable bool `json:"-"`
}

// Icon is a material design icon to represent the entity. Must be prefixed mdi:. If not provided, default value is mdi:cellphone.
type Icon = string

// Location defines location details of the device.
type Location struct {
	// Altitude is the altitude of the device in meters. Must be greater than 0.
	Altitude *int `json:"altitude,omitempty"`

	// Battery is the percentage of battery the device has left. Must be greater than 0.
	Battery *int `json:"battery,omitempty"`

	// Course is the direction in which the device is traveling, measured in degrees and relative to due north. Must be greater than 0.
	Course *int `json:"course,omitempty"`

	// Gps is the current location as latitude and longitude.
	Gps []float32 `json:"gps" validate:"required,number"`

	// GpsAccuracy defines GPS accuracy in meters. Must be greater than 0.
	GpsAccuracy  int     `json:"gps_accuracy" validate:"required,number,gte=0"`
	LocationName *string `json:"location_name,omitempty"`

	// Speed is the speed of the device in meters per second. Must be greater than 0.
	Speed *int `json:"speed,omitempty"`

	// VerticalAccuracy is the accuracy of the altitude value, measured in meters. Must be greater than 0.
	VerticalAccuracy *int `json:"vertical_accuracy,omitempty"`
}

// Name is a human-friendly name for a entity.
type Name = string

// Sensor defines model for Sensor.
type Sensor struct {
	// Attributes defines additional custom attributes of a entity.
	Attributes Attributes `json:"attributes,omitempty"`

	// DeviceClass is a valid Binary Sensor or Sensor device class.
	DeviceClass *string `json:"device_class,omitempty"`

	// Disabled indicates if the entity should be enabled or disabled.
	Disabled *bool `json:"disabled,omitempty"`

	// EntityCategory is the entity category of the entity.
	EntityCategory *EntityCategory `json:"entity_category,omitempty"`

	// Icon is a material design icon to represent the entity. Must be prefixed mdi:. If not provided, default value is mdi:cellphone.
	Icon *Icon `json:"icon,omitempty"`

	// Name is a human-friendly name for a entity.
	Name Name `json:"name" validate:"required"`

	// Retryable indicates whether requests should be retried when sending this sensor data to Home Assistant.
	Retryable bool `json:"-"`

	// State is the current state of the entity.
	State State `json:"state" validate:"required"`

	// StateClass is the state class of the entity (sensors only).
	StateClass *string `json:"state_class,omitempty"`

	// Type is the type of a sensor entity.
	Type SensorType `json:"type" validate:"required"`

	// UniqueID is a unique identifier for a entity.
	UniqueID UniqueID `json:"unique_id" validate:"required"`

	// UnitOfMeasurement is the unit of measurement for the entity.
	UnitOfMeasurement *Units `json:"unit_of_measurement,omitempty"`
}

// SensorRegistration defines model for SensorRegistration.
type SensorRegistration struct {
	// Attributes defines additional custom attributes of a entity.
	Attributes Attributes `json:"attributes,omitempty"`

	// DeviceClass is a valid Binary Sensor or Sensor device class.
	DeviceClass *string `json:"device_class,omitempty"`

	// Disabled indicates if the entity should be enabled or disabled.
	Disabled *bool `json:"disabled,omitempty"`

	// EntityCategory is the entity category of the entity.
	EntityCategory *EntityCategory `json:"entity_category,omitempty"`

	// Icon is a material design icon to represent the entity. Must be prefixed mdi:. If not provided, default value is mdi:cellphone.
	Icon *Icon `json:"icon,omitempty"`

	// Name is a human-friendly name for a entity.
	Name Name `json:"name" validate:"required"`

	// State is the current state of the entity.
	State State `json:"state" validate:"required"`

	// StateClass is the state class of the entity (sensors only).
	StateClass *string `json:"state_class,omitempty"`

	// Type is the type of a sensor entity.
	Type SensorType `json:"type" validate:"required"`

	// UniqueID is a unique identifier for a entity.
	UniqueID UniqueID `json:"unique_id" validate:"required"`

	// UnitOfMeasurement is the unit of measurement for the entity.
	UnitOfMeasurement *Units `json:"unit_of_measurement,omitempty"`
}

// SensorState defines the current state of a sensor.
type SensorState struct {
	// Attributes defines additional custom attributes of a entity.
	Attributes Attributes `json:"attributes,omitempty"`

	// Icon is a material design icon to represent the entity. Must be prefixed mdi:. If not provided, default value is mdi:cellphone.
	Icon *Icon `json:"icon,omitempty"`

	// State is the current state of the entity.
	State State `json:"state" validate:"required"`

	// Type is the type of a sensor entity.
	Type SensorType `json:"type" validate:"required"`

	// UniqueID is a unique identifier for a entity.
	UniqueID UniqueID `json:"unique_id" validate:"required"`
}

// SensorType is the type of a sensor entity.
type SensorType string

// State is the current state of the entity.
type State = interface{}

// UniqueID is a unique identifier for a entity.
type UniqueID = string

// Units is the unit of measurement for the entity.
type Units = string

// AsEvent returns the union data inside the Entity as a Event
func (t Entity) AsEvent() (Event, error) {
	var body Event
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromEvent overwrites any union data inside the Entity as the provided Event
func (t *Entity) FromEvent(v Event) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeEvent performs a merge with any union data inside the Entity, using the provided Event
func (t *Entity) MergeEvent(v Event) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

// AsLocation returns the union data inside the Entity as a Location
func (t Entity) AsLocation() (Location, error) {
	var body Location
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromLocation overwrites any union data inside the Entity as the provided Location
func (t *Entity) FromLocation(v Location) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeLocation performs a merge with any union data inside the Entity, using the provided Location
func (t *Entity) MergeLocation(v Location) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

// AsSensor returns the union data inside the Entity as a Sensor
func (t Entity) AsSensor() (Sensor, error) {
	var body Sensor
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromSensor overwrites any union data inside the Entity as the provided Sensor
func (t *Entity) FromSensor(v Sensor) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeSensor performs a merge with any union data inside the Entity, using the provided Sensor
func (t *Entity) MergeSensor(v Sensor) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

func (t Entity) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *Entity) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}
