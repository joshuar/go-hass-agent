// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	ipv4 = 4
	ipv6 = 6
)

type connectionAddrSensor struct {
	bus         *dbusx.Bus
	configProp  *dbusx.Property[dbus.ObjectPath]
	name        string
	id          string
	addr        string
	gateway     string
	nameservers []string
	linux.Sensor
	prefix int
	ver    int
}

func (c *connectionAddrSensor) Name() string {
	return c.name
}

func (c *connectionAddrSensor) ID() string {
	return c.id
}

func (c *connectionAddrSensor) Attributes() map[string]any {
	attributes := c.Sensor.Attributes()
	attributes["prefix"] = c.prefix
	attributes["gateway"] = c.gateway

	if c.ver == ipv4 {
		attributes["nameservers"] = c.nameservers
	}

	return attributes
}

func (c *connectionAddrSensor) State() any {
	return c.addr
}

func (c *connectionAddrSensor) setState(configPath any) error {
	var path string
	// Acceptable values for configPath are converted to a string.
	switch value := configPath.(type) {
	case string:
		path = value
	case dbus.ObjectPath:
		path = string(value)
	case dbus.Variant:
		variant, err := dbusx.VariantToValue[string](value)
		if err != nil {
			return fmt.Errorf("could not transform variant into objectPath: %w", err)
		}

		path = variant
	default:
		return ErrUnsupportedValue
	}

	if err := c.updateAddr(path); err != nil {
		return fmt.Errorf("could not update address: %w", err)
	}

	c.updateGateway(path)
	c.updateNameservers(path)

	return nil
}

func (c *connectionAddrSensor) updateState() error {
	configPath, err := c.configProp.Get()
	if err != nil {
		return fmt.Errorf("cannot update address: %w", err)
	}

	if err := c.setState(configPath); err != nil {
		return fmt.Errorf("cannot update address: %w", err)
	}

	return nil
}

func (c *connectionAddrSensor) updateAddr(path string) error {
	addrIntr := dBusNMObj + ".IP" + strconv.Itoa(c.ver) + "Config.AddressData"

	// Get the address details property using the given config path.
	addrDetails, err := dbusx.NewProperty[[]map[string]dbus.Variant](c.bus, path, dBusNMObj, addrIntr).Get()
	if err != nil {
		return fmt.Errorf("could not retrieve address data from D-Bus: %w", err)
	}

	var (
		address string
		prefix  int
	)

	if len(addrDetails) > 0 {
		address, err = dbusx.VariantToValue[string](addrDetails[0]["address"])
		if err != nil {
			return fmt.Errorf("could not parse address: %w", err)
		}

		prefix, err = dbusx.VariantToValue[int](addrDetails[0]["prefix"])
		if err != nil {
			return fmt.Errorf("could not parse prefix: %w", err)
		}
	}

	c.addr = address
	c.prefix = prefix

	return nil
}

func (c *connectionAddrSensor) updateGateway(configPath string) {
	gatewayIntr := dBusNMObj + ".IP" + strconv.Itoa(c.ver) + "Config.Gateway"

	// Get the gateway property using the given config path.
	gateway, err := dbusx.NewProperty[string](c.bus, configPath, dBusNMObj, gatewayIntr).Get()
	if err != nil {
		slog.With(slog.String("connection", c.name)).Debug("Could not retrieve gateway from D-Bus.", slog.Any("error", err))
	} else {
		c.gateway = gateway
	}
}

func (c *connectionAddrSensor) updateNameservers(path string) {
	if c.ver == ipv6 {
		return
	}

	nameserversIntr := dBusNMObj + ".IP" + strconv.Itoa(c.ver) + "Config.NameserverData"

	// Get the gateway property using the given config path.
	nameservers, err := dbusx.NewProperty[[]map[string]dbus.Variant](c.bus, path, dBusNMObj, nameserversIntr).Get()
	if err != nil {
		slog.With(slog.String("connection", c.name)).Debug("Could not retrieve nameservers from D-Bus.", slog.Any("error", err))
	} else {
		for _, details := range nameservers {
			nameserver, err := dbusx.VariantToValue[string](details["address"])
			if err != nil {
				continue
			}

			c.nameservers = append(c.nameservers, nameserver)
		}
	}
}

func newConnectionAddrSensor(bus *dbusx.Bus, ver int, connectionPath, connectionName string) *connectionAddrSensor {
	var configPropName string

	name := connectionName + " Connection IPv" + strconv.Itoa(ver) + " Address"
	id := strcase.ToSnake(connectionName) + "_connection_ipv" + strconv.Itoa(ver) + "_address"
	icon := "mdi:numeric-" + strconv.Itoa(ver)

	switch ver {
	case ipv4:
		configPropName = connectionIPv4ConfigProp
	case ipv6:
		configPropName = connectionIPv6ConfigProp
	}

	return &connectionAddrSensor{
		bus:        bus,
		ver:        ver,
		name:       name,
		id:         id,
		Sensor:     linux.Sensor{SensorSrc: linux.DataSrcDbus, IconString: icon},
		configProp: dbusx.NewProperty[dbus.ObjectPath](bus, connectionPath, dBusNMObj, configPropName),
	}
}
