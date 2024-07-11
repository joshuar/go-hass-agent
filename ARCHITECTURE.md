# Architecture

## 📔 Table of Contents

- [Architecture](#architecture)
  - [📔 Table of Contents](#-table-of-contents)
  - [Project Layout](#project-layout)
  - [internal/device](#internaldevice)

## Project Layout

- *build*: build related files and tooling.
  - *magefiles*: code for mage build system.
  - *package*: files related to packaging.
  - *scripts*: misc scripts used by mage and the ci pipeline.
- *deployments*: local storage for the Home Assistant and Mosquitto containers.
- *dist*: temporary storage of artifacts created during ci build/release actions.
- *init*: init scripts for running the agent.
- *internal*: internal agent code.
  - *agent*: code relating to the agent itself.
  - *commands*: code for custom MQTT commands.
  - *device*: device (OS-agnostic) code.
  - *hass*: code relating to working with Home Assistant.
  - *linux*: code for linux specific sensors and control.
  - *logging*: logging code.
  - *preferences*: code for agent preferences.
  - *scripts*: code for custom script sensors.
  - *translations*: code for providing translations in the UI.
- *pkg*: code that could be used by other projects and packages.
  - *dbusx*: code for extending godbus.
  - *hwmon*: code for retrieving sensors from the Linux hwmon interface.
  - *proc*: code for retrieving sensors from the Linux `/proc/` filesystem.
  - *pulseaudio*: code for working with Pulseaudio.
  - *whichdistro*: code for parsing `/etc/os-release` and retrieving its values.

## internal/device

- Mainly code for representing a device running the agent as a `hass.DeviceInfo`.
- Contains some OS-agnostic sensors such as external IP and agent version.
