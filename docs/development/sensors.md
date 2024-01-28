<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Device Sensors

## Code Location

Platform/device sensor code belongs under `GOARCH/`. Ideally, create a new `.go`
file for each sensor or sensor group you are adding.

## Representing a sensor

For sensor data to be registered and sent to Home Assistant, it needs to meet
implement the `tracker.Sensor` interface. It should satisfy the following
methods.

```go
type SensorUpdate interface {
  Name() string
  ID() string
  Icon() string
  SensorType() SensorType
  DeviceClass() SensorDeviceClass
  StateClass() SensorStateClass
  State() interface{}
  Units() string
  Category() string
  Attributes() interface{}
}
```

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

The [Material Design Icon](https://pictogrammers.github.io/@mdi/font/2.0.46/)
representing the current state. It can be changed dynamically based on the
current state or remain constant. Format is “mdi:icon_name”.

### SensorType() hass.SensorType

The `sensor.SensorType` for this sensor. Either `TypeSensor` or `TypeBinary`.

### DeviceClass() hass.SensorDeviceClass

The `sensor.SensorDeviceClass` for this sensor. There are many, pick an
appropriate one (see 
[`internal/hass/sensor/deviceClass.go`](../../internal/hass/sensor/deviceClass.go))
for the sensor or return a value of 0 to indicate it has none.

### StateClass() hass.SensorStateClass

The `sensor.SensorStateClass` for this sensor. There are a few, pick an
appropriate one for the sensor (see
[`internal/hass/sensor/stateClass.go`](../../internal/hass/sensor/stateClass.go))
sensor return a value of 0 to indicate it has none.

### State() interface{}

The current state (value) of the sensor. For Home Assistant, a valid return type
would be `bool` (for `TypeBinary`), `float`, `int`, or `string`.

### Units() string

What units the state should be represented as. If you have defined a
`SensorDeviceClass` for this sensor, that will likely dictate the units you should
use.

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
func SensorUpdater(context.Context) chan tracker.Sensor
```

- The `context.Context` parameter is a context which you should respect a
  potential cancellation of (see below). It can also contain any device specific
  context values, such as common API configuration data.
- The `tracker.Sensor` return value is a channel of sensor values to be updated.

Within this function, you should create the sensors you want to report to Home
Assistant and set up a way to send updates. You will want:

- A way to get the sensor data you need. Most likely stored in a struct.
- Ensure this data struct satisfies the `tracker.Sensor` interface requirements.

How this is achieved will vary and can be done in any way. For example, the
battery sensor data is manifested by listening for D-Bus signals that indicate a
battery property has changed.

You can create as many sensors as you require.

Create a `chan tracker.Sensor` and return this from the updater function. Then,
whenever you have a sensor update, send it via the created channel.

As mentioned you should respect/expect cancellation of the context received as a
parameter. You can do this by including code similar to the following in your function:

```go
go func() {
  // Likely you'll want to clean up the sensor channel...
  // defer close(chan tracker.Sensor)
  <-ctx.Done
  // any additional clean up code can go here...
}
```

Pseudo Go code of what a complete function would look like:

```go
func SensorUpdater(ctx context.Context) chan tracker.Sensor {
  sensorCh := make(chan tracker.Sensor, 1)
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

To have your sensors tracked by the agent, you should add the `SensorUpdater`
function to the `device.SensorInfo` struct. You can do this in the following
function, that should be created for each operating system:

```go
func SetupSensors() *SensorInfo {
  sensorInfo := NewSensorInfo()
  sensorInfo.Add("Some name", SensorUpdater)
  // add each SensorUpdater function here
  return sensorInfo
}
```

This function will be called once by the agent when setting up its sensor
tracker and will run each `SensorUpdater` function that has been defined.

## Examples

See the `apps_linux.go` or `battery_linux.go` files for examples of code
to track the current/running apps and battery states on Linux. These use D-Bus
events for tracking the changes. That is only one possible way to get the
updates; any other method you can think of would probably work as well.
