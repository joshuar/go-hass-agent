<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Go Hass Agent Development Notes

## Build Requirements

In addition to a Go installation, Go Hass Agent needs a few other Go tools installed:

```shell
# goreleaser is used to create release artifacts
go install github.com/goreleaser/goreleaser@latest
# fyne-cross is used build some artifacts that goreleaser doesn't
go install github.com/fyne-io/fyne-cross@latest
# stringer auto-generates some String() methods
go install golang.org/x/tools/cmd/stringer@latest
# gotext is used for translations
go install golang.org/x/text/cmd/gotext@latest
# moq is used for unit tests
go install github.com/matryer/moq@latest
```

The Fyne UI toolkit that is used by Go Hass Agent also requires some development
libraries installed. See the [Fyne prerequisites](https://developer.fyne.io/started/#prerequisites)
 documentation for what you need to install.

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

## Building

Go Hass Agent makes use of `go generate` to generate some of the code. A typical
build process would be:

```shell
go generate ./...
go build
```

### Packages

Go Hass Agent uses [goreleaser](https://goreleaser.com/intro/) to create
packages for Fedora, Arch, and Ubuntu and
[fyne-cross](https://github.com/fyne-io/fyne-cross) to create packages for
Debian.

To build a “local-only” package with `goreleaser`:

```shell
goreleaser release --snapshot --clean
```

Packages will be available under the `dist/` folder.

See the [goreleaser docs](https://goreleaser.com/quick-start/) for more commands
and information.

To build a package for Debian with `fyne-cross`:

```shell
fyne-cross linux -icon assets/trayicon/logo-pretty.png -release
```

The `.tar.xz` will be available under `fyne-cross/dist/linux-amd64/`.

### Container Images

A Dockerfile that you can use to build an image can be found [here](../../Dockerfile).

You can build an image with a command like the following (using Podman):

```shell
podman build --file ./Dockerfile --network host --tag go-hass-agent
```

## Committing Code

This repository is using [conventional commit
messages](https://www.conventionalcommits.org/en/v1.0.0/#summary). This provides
the ability to automatically include relevant notes in the
[changelog](../CHANGELOG.md). The [TL;DR](https://en.wikipedia.org/wiki/TL;DR)
is when writing commit messages, add a prefix:

- `feat:` for a new feature, like a new sensor.
- `fix:` when fixing an issue.
- `refactor:` when making non-visible but useful code changes.
- …and so on. See the link above or see the existing commit messages for examples.

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
