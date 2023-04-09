# go-hass-app

![MIT](https://img.shields.io/github/license/joshuar/go-hass-agent) 
![GitHub last commit](https://img.shields.io/github/last-commit/joshuar/go-hass-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuar/go-hass-agent?style=flat-square)](https://goreportcard.com/report/github.com/joshuar/go-hass-agent) 
[![Go Reference](https://pkg.go.dev/badge/github.com/joshuar/go-hass-agent.svg)](https://pkg.go.dev/github.com/joshuar/go-hass-agent)
[![Release](https://img.shields.io/github/release/joshuar/go-hass-agent?style=flat-square)](https://github.com/joshuar/go-hass-agent/releases/latest)

A [Home Assistant](https://www.home-assistant.io/), [native app
integration](https://developers.home-assistant.io/docs/api/native-app-integration)
for desktop/laptop devices.

Currently, only Linux is supported. Though the code is designed to be extensible
to other operating systems. See [Agent/Extending](docs/agent/extending.md) for
details on how to extend for other operating systems.

## Features

This app will add some sensors to a Home Assistant instance:

- Device location.
- Current active application and list of running applications.
- Battery status (for example, laptop battery and any peripherals).
- Network status (for example, network connection status and IP addresses and
  wifi details where relevant).

The code is designed to be extended to add additional sensors. See
[Device/Sensors](docs/device/sensors.md) for details on adding additional sensors.

## Use-cases

As examples of some of the things that can be done with the data published by this app:

- Change your lighting depending on what active/running apps are on your
  laptop/desktop. For example, you could set your lights dim or activate a scene
  when you are gaming. 
- With your laptop plugged into a smart plug that is also controlled by Home
  Assistant, turn the smart plug on/off based on the laptop battery charge to
  force a full charge/discharge cycle of the battery, extending its life over
  leaving it constantly charged. 
- Like on mobile devices, create automations based on the location of your
  laptop running this app. 
- Receive notifications from Home Assistant on your desktop/laptop.

See also the [FAQ](docs/faq.md). 

## Contributing

I would welcome your contribution! If you find any improvement or issue you want
to fix, feel free to send a pull request!

### Development

Some notes on the development and architecture can be found in the [`docs`](docs/) folder
in the source repository. This might help anyone looking to contribute, extend
or fork this tool.

### Translations

While this application does not have many points where text is displayed to
the end user (logging aside), translation is supported through the `language`
and `message` packages that are part of
[golang.org/x/text](https://pkg.go.dev/golang.org/x/text). 

I would welcome pull requests for translations!



