<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Go Hass Agent

![Maintained](https://img.shields.io/badge/Maintained%3F-yes-green.svg)
![MIT](https://img.shields.io/github/license/joshuar/go-hass-agent)
![GitHub last
commit](https://img.shields.io/github/last-commit/joshuar/go-hass-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuar/go-hass-agent?style=flat-square)](https://goreportcard.com/report/github.com/joshuar/go-hass-agent)
[![Release](https://img.shields.io/github/release/joshuar/go-hass-agent?style=flat-square)](https://github.com/joshuar/go-hass-agent/releases/latest)

A [Home Assistant](https://www.home-assistant.io/), [native app
integration](https://developers.home-assistant.io/docs/api/native-app-integration)
for desktop/laptop devices.

## 🎉 Features

### 📈 Sensors

This app will add some sensors to a Home Assistant instance:


- Device location.
- Current active application and list of running applications.
- Current active user count (and list of usernames).
- Battery status (for example, laptop battery and any peripherals).
- Network status (for example, network connection status, internal and external
  IP addresses and Wi-Fi details where relevant).
- Network transfer rates.
- Memory and swap usage (total/free/used).
- Disk space usage and IO counts/rates.
- Load Averages.
- Uptime.
- Current power profile (Linux only)
- Screen lock status.
- Hardware sensors (temps, fan speeds, alarm status).
- Problems detected by ABRT (Linux only).
- User-specified [script](docs/scripts.md) output.

A full list of sensors can be found in the [docs](docs/sensors.md).

### 🕹️ Controls (via MQTT)

If you have Home Assistant configured with
[MQTT](https://www.home-assistant.io/integrations/mqtt/), you can set up Go Hass
Agent to respond to commands from Home Assistant. See the [docs](docs/mqtt.md).
A number of power related commands (suspend, poweroff, lock screen), among
others, are built-in and any arbitrary D-Bus call can be called via a service
call in Home Assistant.

### And More…

Check out [what I'm working
on](https://github.com/joshuar/go-hass-agent/discussions/150) for future releases.

## 🤔 Use-cases

As examples of some of the things that can be done with the data published by this app:

- Change your lighting depending on:
  - What active/running apps are on your laptop/desktop. For example, you could set your lights dim or activate a scene
  when you are gaming.
  - Whether your screen is locked or the device is shutdown/suspended.
- With your laptop plugged into a smart plug that is also controlled by Home
  Assistant, turn the smart plug on/off based on the battery charge. This can
  force a full charge/discharge cycle of the battery, extending its life over
  leaving it constantly charged.
- Like on mobile devices, create automations based on the location of your
  laptop running this app.
- Monitor network the data transfer amount from the device, useful where network
  data might be capped.
- Monitor CPU load, disk usage and any temperature sensors emitted from the device.
- Receive notifications from Home Assistant on your desktop/laptop. Potentially
  based on or utilising any of the data above.

See also the [FAQ](docs/faq.md).

## 🤝 Compatibility

Currently, only Linux is supported. Though the code is designed to be extensible
to other operating systems. See development information in the
[docs](docs/README.md) for details on how to extend for other operating systems.

## ⬇️ Installation

### 📦 Packages

Head over to the [releases](https://github.com/joshuar/go-hass-agent/releases)
page and download the appropriate package for your operating system and/or
distribution:

- **Fedora**: use the `.rpm`.
- **Ubuntu**: use the `.deb`.
- **Debian**: use the `.tar.xz`.
- **Arch**: use the `.tar.zst`.

Packages (and binaries) are available for **amd64**, **arm** and **arm64**
architectures.

For distributions not listed above, you can try the binary, or build it
yourself from source (see development [docs](docs/README.md)). Note that while
Go is known for statically compiled binaries that “run anywhere”, the Fyne UI
toolkit used by Go Hass Agent makes use of shared libraries that may need to
be installed as well.

Package signatures can be verified with
[cosign](https://github.com/sigstore/cosign). To verify a package, you'll need
to download [cosign.pub](cosign.pub) public key and the `.sig` file (downloaded from
[releases](https://github.com/joshuar/go-hass-agent/releases)) that matches the
package you want to verify. To verify a package, a command similar to the
following for the `rpm` package can be used:

```shell
cosign verify-blob --key cosign.pub --signature go-hass-agent-*.rpm.sig go-hass-agent-*.rpm
```

### 🚢 Container

Container images are available on
[ghcr.io](https://github.com/joshuar/go-hass-agent/pkgs/container/go-hass-agent).
Note that it is recommended to use an image tagged with the latest release version over
the latest container image, which might be unstable.

## 🖱️ Usage

Go Hass Agent runs as a tray icon by default. It is operating system,
distribution and desktop-environment agnostic and should manifest itself in any
tray of any desktop environment.

### First-run

On first-run, Go Hass Agent will display a window where you will need to enter
some details, so it can register itself with a Home Assistant instance to be
able to report sensors and receive notifications.

![Registration Window](assets/screenshots/registration.png)

**You will need:**

- A long-lived access token. You can generate one on your [account profile
  page](https://www.home-assistant.io/docs/authentication/#your-account-profile).
- The web address (URL) on which a Home Assistant instance can be found.
  - Go Hass Agent will try to auto-detect this for you, and you can select it in
    the _Auto-discovered servers_ list. Otherwise, you will need to select _Use
    Custom Server?_, and enter the details manually in _Manual Server Entry_.

When you have entered all the details, click **Submit** and the agent should
start running and reporting sensors to the Home Assistant instance.

### Running “Headless”

Go Hass Agent will automatically detect if there is no GUI available and run in
a “headless” mode with no UI. Registration will need to be completed manually as
a first step in such environments.

You can register Go Hass Agent on the command-line with by
running:

```shell
go-hass-agent --terminal register --token _TOKEN_ --server _URL_
```

You will need to provide a long-lived token `_TOKEN_` and the URL of your Home
Assistant instance, `_URL_`.

Once registered, running Go Hass Agent again with no options should start
tracking and sending sensor data to Home Assistant.

If desired, headless mode can be forced, even in graphical environments, by
specifying the `--terminal` command-line option.

If you want to run Go Hass Agent as a service on a headless machine, see the
[FAQ](docs/faq.md#q-i-want-to-run-the-agent-on-a-server-as-a-service-without-a-gui-can-i-do-this). 

### Running in a container

There is rough support for running Go Hass Agent within a container. Pre-built
images [are available](https://github.com/joshuar/go-hass-agent/pkgs/container/go-hass-agent).

To register the agent running in a container, run the following:

```shell
podman run --rm --hostname go-hass-agent-container \
  --network host \
  --volume go-hass-agent:/home/gouser \
  ghcr.io/joshuar/go-hass-agent register \
  --server https://some.server:port \
  --token longlivedtoken
```

Adjust the `--server` and `--token` values as appropriate.

Once registered, run the agent with:

```shell
podman run --hostname go-hass-agent-container --name my-go-hass-agent \
  --network host \
  --volume go-hass-agent:/home/gouser \
  --volume /proc:/host/proc:ro --volume /sys:/host/sys:ro \
  --volume /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket:ro \
  --volume /run/user/1000/bus:/run/user/1000/bus:ro \
  ghcr.io/joshuar/go-hass-agent run
```

Change the value passed to `--name` to a unique name for your running container
and `--hostname` for the hostname that will be presented to Home Assistant
during registration.

All the other volume mounts are optional, but functionality and the sensors
reported will be severely limited without them.

### Regular Usage

When running, Go Hass Agent will appear as a device under the Mobile App
integration in your Home Assistant instance. It should also report a list of
sensors/entities you can use in any automations, scripts, dashboards and other
parts of Home Assistant.

[![Open your Home Assistant instance to the mobile_app integration.](https://my.home-assistant.io/badges/integration.svg)](https://my.home-assistant.io/redirect/integration/?domain=mobile_app)

## Issues, Feature Requests, Contributing

- Found an issue? Please [report
  it](https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D)!
- Have a suggestion for a feature? Want a particular sensor/measurement added?
  Submit a [feature
  request](https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=feature_request.md&title=)!
- Want to help develop Go Hass Agent? See the [contributing
  guidelines](CONTRIBUTING.md).

## Versioning

This project follows [semantic versioning](https://semver.org/). Given a version number
`MAJOR`.`MINOR`.`PATCH`, the gist of it is:

- A `MAJOR` number change means breaking changes from the previous release.
- A `MINOR` number change means significant changes and new features have been added, but not breaking changes. 
- A `PATCH` number change indicate minor changes and bug fixes.

## 🙌 Acknowledgements

The app icon is taken from the [Home Assistant
project](https://github.com/home-assistant/assets).

## License

[MIT](LICENSE)
