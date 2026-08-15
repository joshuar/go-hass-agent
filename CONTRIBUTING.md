<!--
 Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
 SPDX-License-Identifier: MIT
-->

## Table of Contents

- [🧑‍🤝‍🧑 Contributing](#-contributing)
  - [Development Contributions](#development-contributions)
    - [💾 Committing Code](#-committing-code)
  - [Building From Source](#building-from-source)
    - [Build Requirements](#build-requirements)
    - [Compiling](#compiling)
    - [Cross Compilation](#cross-compilation)
    - [Packages](#packages)
    - [Container Images](#container-images)
  - [Architecture](#architecture)
    - [Terminology](#terminology)
    - [Code Layout](#code-layout)
    - [Execution](#execution)

# 🧑‍🤝‍🧑 Contributing

Thanks for taking an interest in Go Hass Agent! There are **lots** of ways you can contribute to this project:

- Helping with development.
- Helping with translations.
- Just using the agent and providing feedback.

## Development Contributions

I would welcome your contribution! If you find any improvement or issue you want
to fix, feel free to send a pull request!

Some documentation for development can be found in
the [docs](docs/README.md). There is information for developing
Go Hass Agent for different operating systems as well as adding additional
sensors. This might help anyone to look to contribute, extend or fork this tool.

### 💾 Committing Code

This repository is using [conventional commit messages](https://www.conventionalcommits.org/en/v1.0.0/#summary). This
provides the ability to automatically include relevant notes in the [changelog](CHANGELOG.md). The
[TL;DR](https://en.wikipedia.org/wiki/TL;DR) is when writing commit messages, add a prefix:

- `feat:` for a new feature, like a new sensor.
- `fix:` when fixing an issue.
- `refactor:` when making non-visible but useful code changes.
- …and so on. See the link above or see the existing commit messages for
  examples.

## Building From Source

### Build Requirements

Besides Go, Go Hass Agent requires a javascript runtime/toolkit to bundle/build some assets required for the web UI.
[Nodejs](https://nodejs.org/en) works just fine, is packaged in nearly all distributions and has good cross-platform
support.

> [!NOTE]
>
> The devcontainer has all the necessary tooling installed for building Go Hass Agent.

### Compiling

From the root of the Go Hass Agent repository, use the following commands will
build/bundle everything needed:

```shell
npm install
npm run build:js
npm run build:css
# the -X ... linker option is *required*
CGO_ENABLED=0 go build -ldflags="-w -s -X github.com/joshuar/go-hass-agent/config.AppVersion=$(git describe --tags --always --long --dirty)" -o dist/go-hass-agent
```

This will build a binary and place it in `dist/go-hass-agent`.

[⬆️ Back to Top](#table-of-contents)

### Cross Compilation

Go Hass Agent can also be built for **arm (v6/v7)** and **arm64** with cross-compilation. Just change the `go build` in
the commands above as appropriate. For e.g.:

```shell
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X github.com/joshuar/go-hass-agent/config.AppVersion=$(git describe --tags --always --long --dirty)" -o dist/go-hass-agent
```

[⬆️ Back to Top](#table-of-contents)

### Packages

Go Hass Agent uses [nfpm](https://nfpm.goreleaser.com/) to create packages for Fedora, Arch, and Ubuntu/Debian.

To build packages, use the following invocations:

```shell
for format in rpm deb archlinux; do
  go run github.com/goreleaser/nfpm/v2/cmd/nfpm@latest package --packager ${format} --config .nfpm.yaml --target dist
done
```

This will build packages for all possible formats and they will be available under the `dist/` folder.

[⬆️ Back to Top](#table-of-contents)

### Container Images

A Dockerfile that you can use to build an image can be found [here](../../Dockerfile).

You can build an image with a command like the following (using Podman):

```shell
podman build --file ./Dockerfile --tag go-hass-agent
```

As with building a binary,
[cross-compliation](https://docs.docker.com/build/building/multi-platform/#cross-compilation)
is supported:

```shell
# use either linux/arm64, linux/arm/v7 or linux/arm/v6
podman build --file ./Dockerfile --platform linux/arm/v7 --tag go-hass-agent
```

> [!NOTE]
>
> By default, the container will run as a user with UID/GID 1000/1000. You can pick a different UID/GID when building by
> adding `--build-arg UID=999` and `--build-arg GID=999` (adjusting the values as appropriate).

[⬆️ Back to Top](#table-of-contents)

## Architecture

### Terminology

I use a bit of terminology in the code:

- *entity_workers*, *mqtt_workers*, *sensor_workers* or just *workers* are all essentially equivalent. A worker is what
  I've called anything that satisfies one of the interfaces in `agent/workers/workers.go`, namely an **EntityWorker** or
  **PollingEntityWorker** These are the objects that generate various related sensors, for example a worker that
  generates load average sensors. That worker satisfies the **PollingEntityWorker** interface, which the agent can
  schedule via the internal quartz scheduler to poll for the load averages on an interval. The screen lock worker on the
  other-hand, satisfies **EntityWorker** and it listens to D-Bus to receive the screen-lock events, that it then sends
  back to the agent when they occur. **MQTTWorker** can be used for workers that expose controls or sensors over MQTT.
- *entity* or *sensor* refer to the entities/sensors in Home Assistant. For example, the 5m load average sensor. Workers
  might produce a bunch of these (like the load average worker). device would be anything that is running the agent. A
  laptop, server, phone, toaster.
- *agent* refers to the agent as a whole itself. The thing that starts up, manages all the workers, mqtt, web interface
  and everything else.

[⬆️ Back to Top](#table-of-contents)

### Code Layout

The current code layout has the following platform-specific entry-points:

- `agent`: contains the general agent code. The only platform specific part in this code would be a `SetupCtx(ctx
  context.Context) context.Context` that the agent calls to load platform-specific data into the context used by
  workers. For example, on Linux this loads things like D-Bus connections used by various sensor wokers. You'd probably
  need to add a macos specific `SetupCtx` here. You'd then also add all your macos sensor workers in a
  `CreateOSEntityWorkers(ctx context.Context) []workers.EntityWorker` and `CreateOSMQTTWorkers(ctx context.Context)
  (workers.MQTTWorker, error)`, which the agent calls to load all the sensor workers it will manage.
  - Fortunately, it should be easy to add appropriate `*_macos.go` files (or whatever is the right string) similar to
    the existing `*_linux.go` for the above functions.
- `device/`: contains device-specific helpers like methods to fetch the hardware info, used for things like assigning a
  unique ID per device for registering with Home Assistant, determining whether we are running on a desktop, laptop,
  phone, etc. You'd likely need to add macos equivalents of the functions in `info_linux.go`, that would be used by the
  agent to get various device-specific details.
- `pkg/*`: for example, `pkg/linux` contains general platform helpers that aren't tied to a specific sensor and might be
  used by several. For example, there are some packages for dealing with pipewire, the kernel hwmon API, etc. You might
  need to create a new directory under here and pop your macos shared code/helpers there.
- `platform/*`: contains all the sensor code. Under linux, these are grouped under common areas, like media, cpu, memory
  etc. That's just for easier logical organisation, the groups could be anything. Within each group, you have all the
  sensor workers related to it. Most of the macos sensor code would probably live here under an equivalent top-level
  macos directory.

[⬆️ Back to Top](#table-of-contents)

### Execution

1. Agent fires up, loads up a context with SetupCtx.
2. Agent calls CreateOSEntityWorkers and CreateOSMQTTWorkers (passing the loaded context) to get all workers available
   on this device.
3. Agent then starts the workers, including scheduling all polling workers with the internal quartz scheduler, else just
   calling the workers start function.
4. Agent handles sensor data as it comes in from each worker. All comms between workers and the agent is via channels;
   pretty much all workers run in their own goroutine and the agent listens to a channel on which the worker generates
   and sends sensors.

[⬆️ Back to Top](#table-of-contents)
