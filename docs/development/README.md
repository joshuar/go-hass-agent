<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Go Hass Agent Development Notes

## Build Requirements

Go Hass Agent uses [Mage](https://magefile.org/) for development. Make sure you
follow the instructions on the Mage website to install Mage. If you are using
the devcontainer (see below), this is already installed.


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

Use the following mage invocation in the project root directory:

```shell
mage -v -d build/magefiles -w . build:full amd64
```

This will:

- Run `go mod tidy`.
- Run `go fmt ./...`.
- Run `go generate ./...`.
- Build a binary and place it in `dist/go-hass-agent-amd64`.

To just build a binary, replace `build:full` with `build:fast` in the mage
invocation above.

### Packages

Go Hass Agent uses [nfpm](https://nfpm.goreleaser.com/) to create
packages for Fedora, Arch, and Ubuntu and
[fyne-cross](https://github.com/fyne-io/fyne-cross) to create packages for
Debian and Linux distributions with older libraries.

To build packages, use the following invocations:

```shell
mage -v -d build/magefiles -w . package:nfpm amd64
mage -v -d build/magefiles -w . package:fyneCross amd64
```

The above mage actions will install the necessary tooling for packaging, if
needed. 

Packages built with `nfpm` will be available under the `dist/` folder.
Packages built with `fyne-cross` will be available under `fyne-cross/dist/linux-amd64/`.


### Other Architectures

Go Hass Agent can also be built for **arm** and **arm64** with
cross-compilation. **This is only supported on Ubuntu as the host for
cross-compiles**. To build for a different architecture, first install the
appropriate package dependencies: 

```shell
mage -v -d build/magefiles -w . preps:deps arm # or arm64
```

Then the above commands for building and packaging need to just replace `amd64`
with with either `arm` or `arm64`.

> [!NOTE]
> The devcontainer has all the necessary compilers and libraries
> installed for cross-compilation.

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
