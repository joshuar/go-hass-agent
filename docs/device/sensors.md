<!--
 Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Device Sensors

## Code Location

Platform/device sensor code belongs under `device/`. Ideally, create a new `.go`
file for each sensor or sensor group you are adding.

## Representing a sensor

For sensor data to be registered and sent to Home Assistant, it needs to meet
implement the `hass.SensorUpdate` interface. It should satisfy the following
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

An example would be "My Network Sensor".

### ID() string

A unique ID for the sensor. This is used by Home Assistant to identify this
particular sensor and store its state data in its database. This should be
unique and never change. It should be formatted as snake case and all lowercase.

An example would be "my_network_sensor_1".

### Icon() string

The [Material Design Icon](https://pictogrammers.github.io/@mdi/font/2.0.46/)
representing the current state. It can be changed dynamically based on the
current state or remain constant. Format is "mdi:icon_name". 

### SensorType() hass.SensorType 

The `hass.SensorType` for this sensor. Either `TypeSensor` or `TypeBinary`.


### DeviceClass() hass.SensorDeviceClass

The `hass.SensorDeviceClass` for this sensor. There are many, pick an
[appropriate
one](https://developers.home-assistant.io/docs/core/entity/sensor/#available-device-classes)
for the sensor or return a value of 0 to indicate it has none.

### StateClass() hass.SensorStateClass

The `hass.SensorStateClass` for this sensor. There are a few, pick an
[appropriate
one](https://developers.home-assistant.io/docs/core/entity/sensor/#available-state-classes)
for the sensor or return a value of 0 to indicate it has none.

### State() interface{}

The current state (value) of the sensor. For Home Assistant, a valid return type
would be `bool` (for `TypeBinary`), `float`, `int`, or `string`.

### Units() string

What units the state should be represented as. If you have defined a
SensorDeviceClass for this sensor, that will likely dictate the units you should
use.

### Category() string

This affects how the sensor is displayed in the interface. Generally, return
"diagnostic" for an entity exposing some configuration parameter or diagnostics
of a device or an empty string for anything else.

### Attributes() interface{}

Any additional attributes or state/values you would like to associate with the
sensor. This should be formatted as a `struct{}` that can be marshaled into
valid JSON.  

## Managing sensor updates

To track and send sensor updates to Home Assistant, create a function with the
following signature:

```go
func SensorUpdater(ctx context.Context, updateCh chan interface{})
```

The `ctx` parameter will contain the device/platform specific APIs and
variables. You can retrieve those in the function with:

```go
	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor app state.")
		return
	}
```

See [agent/extending](../agent/extending.md) for more details on creating this
context and loading it with the right information.

The `updateCh` will be the channel you use to send the sensor data that
implements the `hass.SensorUpdate` interface. 

Within this function, you should create the sensors you want to report to Home
Assistant and set up a way to send updates. You will want:

- A way to get the sensor data you need. Most likely stored in a struct.
- Ensure this data struct meets the `hass.SensorUpdate` interface requirements.
- Pass the data struct through the `updateCh` channel as needed where it will be
tracked and sent to Home Assistant by the agent.

How this is achieved will vary and can be done in any way. For example, the
battery sensor data is manifested by listening for DBus signals that indicate a
battery property has changed. 

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
to track the current/running apps and battery states on Linux. These use DBus
events for tracking the changes. That is only one possible way to get the
updates; any other method you can think of would probably work as well. 