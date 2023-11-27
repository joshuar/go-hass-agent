<!--
 Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Go Hass Agent

![MIT](https://img.shields.io/github/license/joshuar/go-hass-agent)
![GitHub last commit](https://img.shields.io/github/last-commit/joshuar/go-hass-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuar/go-hass-agent?style=flat-square)](https://goreportcard.com/report/github.com/joshuar/go-hass-agent)
[![Go Reference](https://pkg.go.dev/badge/github.com/joshuar/go-hass-agent.svg)](https://pkg.go.dev/github.com/joshuar/go-hass-agent)
[![Release](https://img.shields.io/github/release/joshuar/go-hass-agent?style=flat-square)](https://github.com/joshuar/go-hass-agent/releases/latest)

A [Home Assistant](https://www.home-assistant.io/), [native app
integration](https://developers.home-assistant.io/docs/api/native-app-integration)
for desktop/laptop devices.

## 🎉 Features

This app will add some sensors to a Home Assistant instance:

- Device location.
- Current active application and list of running applications.
- Battery status (for example, laptop battery and any peripherals).
- Network status (for example, network connection status, internal and external
  IP addresses and Wi-Fi details where relevant).
- Memory and swap usage (total/free/used).
- Disk usage.
- Load Averages.
- Uptime.
- Power profile.
- Screen lock.
- Problems detected by ABRT.
- User-specified [script](docs/scripts.md) output.

A full list of sensors can be found in the [docs](docs/sensors.md).

The code can be extended to add additional sensors. See the [development docs](docs/development.md) 
for details.

## 🤔 Use-cases

As examples of some of the things that can be done with the data published by this app:

- Change your lighting depending on what active/running apps are on your
  laptop/desktop. For example, you could set your lights dim or activate a scene
  when you are gaming.
- With your laptop plugged into a smart plug that is also controlled by Home
  Assistant, turn the smart plug on/off based on the battery charge to
  force a full charge/discharge cycle of the battery, extending its life over
  leaving it constantly charged.
- Like on mobile devices, create automations based on the location of your
  laptop running this app.
- Receive notifications from Home Assistant on your desktop/laptop.

See also the [FAQ](docs/faq.md).

## 🤝 Compatibility

Currently, only Linux is supported. Though the code is designed to be extensible
to other operating systems. See development information in the
[docs](docs/README.md) for details on how to extend for other operating systems.

## ⬇️ Installation

Head over to the [releases](https://github.com/joshuar/go-hass-agent/releases)
page and download the appropriate package for your operating system and/or
distribution:

- For **Fedora**, use the `.rpm`.
- For **Ubuntu**, use the `.deb`.
- For **Debian**, use the `.tar.xz`.
- For **Arch**, use the `.tar.zst`.

For other distributions not listed above, you can try the binary, or build it
yourself from source (see development [docs](docs/README.md)). Note that while
Go is known for statically compiled binaries that "run anywhere", the Fyne UI
toolkit used by Go Hass Agent makes use of shared libraries that may need to
be installed as well.

Package signatures can be verified with
[cosign](https://github.com/sigstore/cosign). To verify a package, you'll need
the [cosign.pub](cosign.pub) public key and the `.sig` file (downloaded from
[releases](https://github.com/joshuar/go-hass-agent/releases)) that matches the
package you want to verify. To verify a package, a command similar to the
following for the `rpm` package can be used:

```shell
cosign verify-blob --key cosign.pub --signature go-hass-agent-*.rpm.sig go-hass-agent-*.rpm
```

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

### Running "Headless"

As alternative, you can register Go Hass Agent on the command-line with by
running:

```shell
go-hass-agent register --token _TOKEN_ --server _URL_
```

You will need to provide a long-lived token `_TOKEN_` and the URL of your Home
Assistant instance, `_URL_`.

### Running in a container

There is rough support for running Go Hass Agent within a container. A
Dockerfile that you can use to build an image can be found
[here](build/package/Dockerfile). 

You can build an image with a command like the following (using Podman):

```shell
cat Dockerfile | podman build --build-arg version='x.x.x' --network host --tag go-hass-agent -
```

The `--build-arg version=x.x.x` is required and the version should correspond to
a [release](releases/). Once the image is built, first register the agent with
Home Assistant:

```shell
podman run --rm --hostname hass-container --volume /proc:/host/proc:ro --volume /sys:/host/sys:ro --volume /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket:ro --volume /run/user/1000/bus:/run/user/1000/bus:ro --volume ~/go-hass-agent:/home/go-hass-agent:U --network host --userns keep-id go-hass-agent register --server some.server:port --token longlivedtoken
```

Once registered, run the agent with:

```shell
podman run --rm --hostname hass-container --volume /proc:/host/proc:ro --volume /sys:/host/sys:ro --volume /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket:ro --volume /run/user/1000/bus:/run/user/1000/bus:ro --volume ~/go-hass-agent:/home/go-hass-agent:U --network host --userns keep-id go-hass-agent
```

You can change `~/go-hass-agent` to whatever volume you want to use to store the agent config
and registry. The D-Bus volumes are optional, not specifying them will mean
sensors that rely on a D-Bus connection will not show up. 

### Regular Usage

When running, Go Hass Agent will appear as a device under the [Mobile
App](https://www.home-assistant.io/integrations/mobile_app) integration in your
Home Assistant instance. It should also report a list of sensors/entities you
can use in any automations, scripts, dashboards and other parts of Home
Assistant.

## 🧑‍🤝‍🧑 Contributing

### Development

I would welcome your contribution! If you find any improvement or issue you want
to fix, feel free to send a pull request!

Some documentation for development can be found in
the [docs](docs/README.md). There is information for developing
Go Hass Agent for different operating systems as well as adding additional
sensors. This might help anyone to look to contribute, extend or fork this tool.

### Translations

While this application does not have many points where text is displayed to
the end user (logging aside), translation is supported through the `language`
and `message` packages that are part of
[golang.org/x/text](https://pkg.go.dev/golang.org/x/text).

I would welcome pull requests for translations!

## 🙌 Acknowledgements

The app icon is taken from the [Home Assistant
project](https://github.com/home-assistant/assets).

## License

[MIT](LICENSE)
