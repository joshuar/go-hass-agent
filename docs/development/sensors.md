<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Device Sensors

## Code Location

Platform/device sensor code belongs under `GOARCH/`. Ideally, create a new `.go`
file for each sensor or sensor group you are adding.

## Representing data as a sensor

For data to be sent to Home Assistant, it needs to implement a few interfaces.

Firstly, the data should satisfy the `sensor.SensorState` interface, which represents the minimum data required to be sent to Home Assistant when the data changes:

```go
type SensorState interface {
	ID() string
	Icon() string
	State() any
	SensorType() types.SensorClass
	Units() string
	Attributes() any
}
```

Initially however, the sensor will need to be registered with Home Assistant and needs to satisfy the `sensor.SensorRegistration` interface:

```go
type SensorRegistration interface {
	SensorState
	Name() string
	DeviceClass() types.DeviceClass
	StateClass() types.StateClass
	Category() string
}
```

The recommended approach is to define your own custom type (usually a `struct`) that satisfies both of these interfaces.

An explanation of the methods follows.

### Name() string

A friendly name for this sensor. It will be what is shown in the Home Assistant
UI for the sensor.

An example would be “My Network Sensor”.

### ID() string

A unique ID for the sensor. This is used by Home Assistant to identify this
particular sensor and store its state data in its database. This should be
unique and never change. It should be formatted as snake case and all lowercase.

An example would be “my_network_sensor_1”.

### Icon() string

The [Material Design Icon](https://pictogrammers.com/library/mdi/)
representing the current state. It can be changed dynamically based on the
current state or remain constant. Format is “mdi:icon_name”.

### SensorType() types.SensorClass

The type of sensor. This is either `types.Sensor` (mapping to a [Home Assistant
Sensor Entity](https://developers.home-assistant.io/docs/core/entity/sensor)) or
`types.Binary` (mapping to a [Home Assistant Binary Sensor
Entity](https://developers.home-assistant.io/docs/core/entity/binary-sensor)).

### DeviceClass() types.DeviceClass

The device class for this sensor, which reflects the type of measurement it is
recording. This maps to the [Home Assistant Device
Classes](https://developers.home-assistant.io/docs/core/entity/sensor#available-device-classes),
such as “temperature”, “data_size”, “speed” etc. See also [source
code](../../internal/hass/sensor/types/DeviceClass.go).

### StateClass() types.StateClass

For the values, what kind of state they represent, such as an instataneous
measurement, an increasing/decreasing value, etc. This maps to the [Home
Assistant State
Classes](https://developers.home-assistant.io/docs/core/entity/sensor#available-state-classes).
See also [source code](../../internal/hass/sensor/types/StateClass.go).

### State() interface{}

The current state (value) of the sensor. For Home Assistant, a valid return type
would be `bool` (for `TypeBinary`), `float`, `int`, or `string`.

### Units() string

What units the state should be represented as. Generally, if this sensor has a
[device class](#deviceclass-typesdeviceclass), that will define the units. 

### Category() string

This affects how the sensor is displayed in the interface. Generally, return
“diagnostic” for an entity exposing some configuration parameter or diagnostics
of a device or an empty string for anything else.

### Attributes() interface{}

Any additional attributes or state/values you would like to associate with the
sensor. This should be formatted as a `struct{}` that can be marshalled into
valid JSON.

## Managing sensor updates

To track and send sensor updates to Home Assistant, create a function with the
following signature:

```go
func SensorUpdater(context.Context) chan sensor.Details
```

- The `context.Context` parameter is a context which you should respect a
  potential cancellation of (see below). It can also contain any device specific
  context values, such as common API configuration data.
- The `sensor.Details` is a catch-all interface that could be either a
  `sensor.SensorState` or `sensor.SensorRegistration`. If you custom type
  satisfies those interfaces, it can be passed out through this channel.

Within this function, you should create the sensors you want to report to Home
Assistant and set up a way to send updates. How this is achieved will vary and
can be done in any way. For example, the battery sensor data for Linux is
manifested by listening for D-Bus signals that indicate a battery property has
changed.

You can create as many sensors as you require. Whenever you have a sensor
update, send it via the `sensor.Details` channel.

As mentioned you should respect/expect cancellation of the context received as a
parameter. You can do this by including code similar to the following in your function:

```go
go func() {
  // Likely you'll want to clean up the sensor channel...
  // defer close(chan sensor.Sensor)
  <-ctx.Done
  // any additional clean up code can go here...
}
```

Pseudo Go code of what a complete function would look like:

```go
func SensorUpdater(ctx context.Context) chan sensor.Details {
  sensorCh := make(chan sensor.Details, 1)
  ...code to set up your sensors...
  for ...some timer, event channel, other loop... {
    ...code to create a sensor object...
    // send your sensor updates
    sensorCh <- sensor
  }
  go func() {
    defer close(sensorCh)
    <-ctx.Done()
  }
}
```

### Helper Functions

There are some helper functions that might be useful to you in
`internal/device/helpers`. For example, the `PollSensors` function can be used
to update sensors on an interval. It adds a bit of jitter to your
interval as well to avoid any “thundering herd” problems.

## Sensor tracking

To have your sensors tracked by the agent, the updater function needs to be
added to the list of these functions returned by the platform-specific
`sensorWorkers()` function. It has the following signature:

```go
sensorWorkers() []func(context.Context) chan sensor.Details
```

An example for Linux can be found in the [source
code](../../internal/agent/device_linux.go). This is called by the agent code to
retrieve the list of updater functions, that the agent then executes in its own
goroutine. Each updater function then runs for the lifetime of the agent,
sending updates via the `sensor.Details` channel back to the agent, which
handles tracking internally and forwarding to Home Assistant.


## Examples

### Linux

- A reusable, common struct is implemented that satisfies the
  `sensor.SensorState` and `Sensor.SensorRegistration` interfaces in
  [internal/linux/sensor.go](../../internal/linux/sensor.go).
- Each sensor then encapsulates this common struct and uses it as appropriate
  for its needs. As an example see the load averages sensors in
  [internal/linux/cpu/loadavgs.go](../../internal/linux/cpu/loadAvgs.go).
