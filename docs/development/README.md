<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Go Hass Agent Development Notes

## Development Environment

It is recommended to use [Visual Studio Code](https://code.visualstudio.com/).
This project makes use of a [Devcontainer](https://containers.dev/) to provide
some convenience during development.

[![Open in Dev Containers](https://img.shields.io/static/v1?label=Dev%20Containers&message=Open&color=blue&logo=visualstudiocode)](https://vscode.dev/redirect?url=vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=https://github.com/joshuar/go-hass-agent)

If using Visual Studio Code, you should be prompted when opening your cloned
copy of the code to set up the dev container. The container contains an
installation of Home Assistant and Mosquitto (MQTT broker) that can be used for
testing. They should be started automatically.

- Home Assistant will be listening on <http://localhost:8123>.
- Mosquitto will be listening on <http://localhost:1833>.

Note that while you can also build and run the agent within the container
environment, this will limit what sensors are reported and may even hinder
development of new sensors. As such, it is recommended to build and run outside
the container. You can still connect to Home Assistant running within the
container, as it is exposed as per above.

## Extending the Agent

### Adding OS support

The intention of the agent design is to make it OS-agnostic.

Most OS specific code for fetching sensor data should likely be part of a
`GOARCH` package and using filename suffixes such as `filename_GOOS_GOARCH.go`.
See the files under `linux/` as examples.

For some OSes, you might need some code to initialise or create some data source
or API that the individual sensor fetching code uses. This code should be placed
in `device/`, using filename suffixes such as `filename_GOOS_GOARCH.go`

For example, on Linux, a D-Bus connection is used for a lot of the sensor data gathering.

In such cases, you should pass this through as a value in a context. You can
create the following function for your platform:

```go
SetupContext(ctx context.Context) context.Context
```

It should accept a `context.Context` and derive its own context from this base
that contains the necessary values for the platform. It will be propagated
throughout the code wherever a context is passed and available for retrieval and
use.

An example can be found in `device/device_linux.go`.
