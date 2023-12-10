// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

// Sensor represents an update for a sensor. It reflects the current state
// of the sensor at the point in time it is used. It provides a bridge between
// platform/device and HA implementations of what a sensor is.
//
//go:generate moq -out mock_Sensor_test.go . Sensor
type Sensor interface {
	Name() string
	ID() string
	Icon() string
	SensorType() sensor.SensorType
	DeviceClass() sensor.SensorDeviceClass
	StateClass() sensor.SensorStateClass
	State() interface{}
	Units() string
	Category() string
	Attributes() interface{}
}

func prettyPrintState(s Sensor) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v", s.State())
	if s.Units() != "" {
		fmt.Fprintf(&b, " %s", s.Units())
	}
	return b.String()
}

func marshallSensorState(state Sensor, registered bool) *sensor.SensorState {
	s := &sensor.SensorState{}
	s.StateAttributes = state.Attributes()
	s.Icon = state.Icon()
	s.State = state.State()
	s.Type = marshalClass(state.SensorType())
	s.UniqueID = state.ID()
	s.Registered = registered
	if !s.Registered {
		s.Name = state.Name()
		s.DeviceClass = marshalClass(state.DeviceClass())
		s.StateClass = marshalClass(state.StateClass())
		s.UnitOfMeasurement = state.Units()
		s.EntityCategory = state.Category()
		s.Disabled = false
	}
	return s
}

type ComparableStringer interface {
	comparable
	String() string
}

func returnZero[T any](s ...T) T {
	var zero T
	return zero
}

func marshalClass[C ComparableStringer](class C) string {
	if class == returnZero[C](class) {
		return ""
	} else {
		return class.String()
	}
}

func MergeSensorCh(ctx context.Context, sensorCh ...<-chan Sensor) <-chan Sensor {
	var wg sync.WaitGroup
	out := make(chan Sensor)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan Sensor) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}
	wg.Add(len(sensorCh))
	for _, c := range sensorCh {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
