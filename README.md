# go-hass-app

A [Home
Assistant](https://www.home-assistant.io/),
[native app
integration](https://developers.home-assistant.io/docs/api/native-app-integration)
for desktop/laptop devices.  

## Features

This app will add some sensors to a Home Assistant instance:

- Device location.
- Current active application and list of running applications.
- Battery status (for example, laptop battery and any peripherals).

## Use-cases

Some of the things that can be done with the data published by this app:

- Change your lighting depending on what active/running apps are on your
  laptop/desktop. For example, you could set your lights dim or to some theme
  when you are gaming. 
- With your laptop plugged into a smart plug that is also controlled by Home Assistant, turn
  the smart plug on/off based on the laptop battery charge to force a full charge/discharge cycle of the
  battery, extending its life over leaving it constantly charged. 
- Like on mobile devices, create automations based on the location of your laptop running this app. 
- Receive notifications from Home Assistant.  

See the [FAQ](docs/faq.md)