<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Agent

## Architecture

### Configuration

The agent needs a valid configuration and will prompt the user to help create it
if one isn't found. Configuration is stored in a TOML formatted file. **The
configuration should not be edited manually**.

### Contexts

The agent creates a cancellable context for all the sensor updater functions it
runs. Any platform code should accept this context and handle context
cancellation gracefully (adding their own timeouts, cancellations as needed).

### Data Updates

Data updates are left entirely up to the individual sensors and the platform the
app is running on. For example, on Linux, location and running apps are
updated as often as the relevant information is published on the user's session
D-Bus. More details about how sensor updates work can be found in
[sensors](sensors.md).

### Notifications

The agent supports receiving notifications from Home Assistant (via a [websocket
connection](https://developers.home-assistant.io/docs/api/native-app-integration/notifications#enabling-websocket-push-notifications)).
Notifications are displayed using the [Fyne Notification
API](https://developer.fyne.io/api/v2.4/notification.html).

## Extending

### Adding a new OS

Nearly all operating system specific code should go under a
`internal/${GOHOSTOS}` directory.

All supported operating systems need to have code that satisfies the
`hass.DeviceInfo` interface:

```go
type DeviceInfo interface {
  DeviceID() string
  AppID() string
  AppName() string
  AppVersion() string
  DeviceName() string
  Manufacturer() string
  Model() string
  OsName() string
  OsVersion() string
  SupportsEncryption() bool
  AppData() interface{}
}
```

This interface reflects the data necessary to register a new native app in Home
Assistant. It is used for initial registration of the agent to Home Assistant.
See the Home Assistant
[documentation](https://developers.home-assistant.io/docs/api/native-app-integration/setup#registering-the-device-1)
for the method values.

For the agent, there needs to be code that satisfies the `agent.Device`
interface:

```go
type Device interface {
  DeviceName() string
  DeviceID() string
  Setup(context.Context) context.Context
}
```

Two of these methods are reused from the `hass.DeviceInfo` interface, so it makes
sense to have a single concrete type (such as a struct) that satisfies both
interfaces.

`Setup(context.Context)` can be used to run any code to set up a device, such as
initialising any APIs. By passing through a context, you can add context values
containing any data that will be needed by the operating system code and return
the new context. 

On Linux, this function creates and then stores the D-Bus connections in a
context value for later use by sensor code. The
[internal/linux/device.go](../../internal/linux/device.go) contains the code as
an example.

Ideally the operating system code would expose a `NewDevice` function that
returns a struct that satisfies the above interfaces. That function should then
be called by operating specific code in the agent package. See the
[internal/agent/device_linux.go](../../internal/agent/device_linux.go) file and
the `newDevice(context.Context)` function for how this works on Linux. The
`newDevice` function is always called before starting sensor updater functions.
By using a file named `*_${GOHOSTOS}.go` containing this code, the `newDevice`
function always does the right thing on each operating system.
