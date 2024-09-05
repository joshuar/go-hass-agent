// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate stringer -type=connState,connIcon -output connectionState_generated.go -linecomment
package net

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	connUnknown      connState = iota // Unknown
	connActivating                    // Activating
	connOnline                        // Online
	connDeactivating                  // Deactivating
	connOffline                       // Offline
)

const (
	iconUnknown      connIcon = iota // mdi:help-network
	iconActivating                   // mdi:plus-network
	iconOnline                       // mdi:network
	iconDeactivating                 // mdi:network-minus
	iconOffline                      // mdi:network-off
)

type connState uint32

type connIcon uint32

type connectionStateSensor struct {
	stateProp *dbusx.Property[connState]
	linux.Sensor
	value connState
}

func (c *connectionStateSensor) Icon() string {
	return connIcon(c.value).String()
}

func (c *connectionStateSensor) State() any {
	return c.value.String()
}

func (c *connectionStateSensor) setState(state any) error {
	switch value := state.(type) {
	case dbus.Variant:
		if state, err := dbusx.VariantToValue[connState](value); err != nil {
			return fmt.Errorf("could not parse updated connection state: %w", err)
		} else {
			c.value = state
		}
	case uint32:
		c.value = connState(value)
	default:
		return ErrUnsupportedValue
	}

	return nil
}

func (c *connectionStateSensor) updateState() error {
	state, err := c.stateProp.Get()
	if err != nil {
		return fmt.Errorf("cannot update state: %w", err)
	}

	c.value = state

	return nil
}

func newConnectionStateSensor(bus *dbusx.Bus, connectionPath, connectionName string) *connectionStateSensor {
	return &connectionStateSensor{
		Sensor: linux.Sensor{
			DataSource:  linux.DataSrcDbus,
			DisplayName: connectionName + " Connection State",
			UniqueID:    strcase.ToSnake(connectionName) + "_connection_state",
		},
		stateProp: dbusx.NewProperty[connState](bus, connectionPath, dBusNMObj, connectionStateProp),
	}
}
