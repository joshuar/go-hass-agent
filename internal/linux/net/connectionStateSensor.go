// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go run golang.org/x/tools/cmd/stringer -type=connState,connIcon -output connectionState_generated.go -linecomment
package net

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
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
	name      string
	state     string
	icon      string
	stateProp *dbusx.Property[connState]
}

func (c *connectionStateSensor) generateEntity(ctx context.Context) (models.Entity, error) {
	return sensor.NewSensor(ctx,
		sensor.WithName(c.name+" Connection State"),
		sensor.WithID(strcase.ToSnake(c.name)+"_connection_state"),
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
		sensor.WithState(c.state),
		sensor.WithIcon(c.icon),
	)
}

func (c *connectionStateSensor) setState(state any) error {
	switch value := state.(type) {
	case dbus.Variant:
		if state, err := dbusx.VariantToValue[connState](value); err != nil {
			return fmt.Errorf("could not parse updated connection state: %w", err)
		} else {
			c.state = state.String()
			c.icon = connIcon(state).String()
		}
	case uint32:
		c.state = connState(value).String()
		c.icon = connIcon(value).String()
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

	c.state = state.String()
	c.state = connIcon(state).String()

	return nil
}

func newConnectionStateSensor(bus *dbusx.Bus, connectionPath, connectionName string) (*connectionStateSensor, error) {
	conn := &connectionStateSensor{
		name:      connectionName,
		stateProp: dbusx.NewProperty[connState](bus, connectionPath, dBusNMObj, connectionStateProp),
	}

	if err := conn.updateState(); err != nil {
		return nil, fmt.Errorf("cannot create connection sensor: %w", err)
	}

	return conn, nil
}
