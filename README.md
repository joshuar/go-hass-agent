<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<div align="center">

  <img src="internal/agent/ui/assets/go-hass-agent.png" alt="logo" width="200" height="auto" />
  <h1>Go Hass Agent</h1>

  <p>
    A <a href="https://www.home-assistant.io/">Home Assistant</a>, <a href="https://developers.home-assistant.io/docs/api/native-app-integration">native app
    integration</a> for desktop/laptop devices.
  </p>

<!-- Badges -->
<p>
  <a href="https://github.com/joshuar/go-hass-agent/graphs/contributors">
    <img src="https://img.shields.io/github/contributors/joshuar/go-hass-agent" alt="contributors" />
  </a>
  <a href="">
    <img src="https://img.shields.io/github/last-commit/joshuar/go-hass-agent" alt="last update" />
  </a>
  <a href="https://github.com/joshuar/go-hass-agent/network/members">
    <img src="https://img.shields.io/github/forks/joshuar/go-hass-agent" alt="forks" />
  </a>
  <a href="https://github.com/joshuar/go-hass-agent/stargazers">
    <img src="https://img.shields.io/github/stars/joshuar/go-hass-agent" alt="stars" />
  </a>
  <a href="https://github.com/joshuar/go-hass-agent/issues/">
    <img src="https://img.shields.io/github/issues/joshuar/go-hass-agent" alt="open issues" />
  </a>
  <a href="https://github.com/joshuar/go-hass-agent/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/joshuar/go-hass-agent.svg" alt="license" />
  </a>
</p>

<h4>
    <a href="https://github.com/joshuar/go-hass-agent">Documentation</a>
  <span> · </span>
    <a href="https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D">Report Bug</a>
  <span> · </span>
    <a href="https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=feature_request.md&title=">Request Feature</a>
  </h4>
</div>

<br />

<!-- Table of Contents -->
- [:star2: About the Project](#star2-about-the-project)
  - [:dart: Features](#dart-features)
    - [📈 Sensors](#-sensors)
    - [🕹️ Controls (via MQTT)](#️-controls-via-mqtt)
  - [🤔 Use-cases](#-use-cases)
  - [Versioning](#versioning)
- [:toolbox: Getting Started](#toolbox-getting-started)
  - [🤝 Compatibility](#-compatibility)
  - [:gear: Installation](#gear-installation)
    - [📦 Packages](#-packages)
    - [🚢 Container](#-container)
- [:eyes: Usage](#eyes-usage)
  - [First-run](#first-run)
  - [Running “Headless”](#running-headless)
  - [Running in a container](#running-in-a-container)
  - [Regular Usage](#regular-usage)
  - [Script Sensors](#script-sensors)
    - [Requirements](#requirements)
    - [Supported Scripting Languages](#supported-scripting-languages)
    - [Output Format](#output-format)
      - [Examples](#examples)
        - [JSON](#json)
        - [YAML](#yaml)
        - [TOML](#toml)
    - [Schedule](#schedule)
    - [Security Implications](#security-implications)
  - [Control via MQTT](#control-via-mqtt)
    - [Requirements](#requirements-1)
    - [Configuration](#configuration)
    - [Available Controls](#available-controls)
    - [Arbitrary D-BUS commands](#arbitrary-d-bus-commands)
    - [Security Implications](#security-implications-1)
- [:compass: Roadmap](#compass-roadmap)
- [:wave: Contributing](#wave-contributing)
  - [:scroll: Code of Conduct](#scroll-code-of-conduct)
- [:grey\_question: FAQ](#grey_question-faq)
- [:gem: Acknowledgements](#gem-acknowledgements)
- [:warning: License](#warning-license)

<!-- About the Project -->
## :star2: About the Project

<!-- Features -->
### :dart: Features

#### 📈 Sensors

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

#### 🕹️ Controls (via MQTT)

If you have Home Assistant configured with
[MQTT](https://www.home-assistant.io/integrations/mqtt/), you can set up Go Hass
Agent to respond to commands from Home Assistant. See the [docs](docs/mqtt.md).
A number of power related commands (suspend, poweroff, lock screen), among
others, are built-in and any arbitrary D-Bus call can be called via a service
call in Home Assistant.

### 🤔 Use-cases

As examples of some of the things that can be done with the data published by
this app:

- Change your lighting depending on:
  - What active/running apps are on your laptop/desktop. For example, you could
  set your lights dim or activate a scene when you are gaming.
  - Whether your screen is locked or the device is shutdown/suspended.
- With your laptop plugged into a smart plug that is also controlled by Home
  Assistant, turn the smart plug on/off based on the battery charge. This can
  force a full charge/discharge cycle of the battery, extending its life over
  leaving it constantly charged.
- Like on mobile devices, create automations based on the location of your
  laptop running this app.
- Monitor network the data transfer amount from the device, useful where network
  data might be capped.
- Monitor CPU load, disk usage and any temperature sensors emitted from the
  device.
- Receive notifications from Home Assistant on your desktop/laptop. Potentially
  based on or utilising any of the data above.

### Versioning

This project follows [semantic versioning](https://semver.org/). Given a version
number `MAJOR`.`MINOR`.`PATCH`, the gist of it is:

- A `MAJOR` number change means breaking changes from the previous release.
- A `MINOR` number change means significant changes and new features have been
  added, but not breaking changes.
- A `PATCH` number change indicate minor changes and bug fixes.

<!-- Getting Started -->
## :toolbox: Getting Started

### 🤝 Compatibility

Currently, only Linux is supported. Though the code is designed to be extensible
to other operating systems. See development information in the
[docs](docs/README.md) for details on how to extend for other operating systems.

<!-- Installation -->
### :gear: Installation

#### 📦 Packages

Head over to the [releases](https://github.com/joshuar/go-hass-agent/releases)
page and download the appropriate package for your operating system and/or
distribution:

- **Fedora**: use the `.rpm`.
- **Ubuntu**: use the `.deb`.
- **Debian**: use the `.tar.xz`.
- **Arch**: use the `.tar.zst`.

Packages (and binaries) are available for **amd64**, **arm (v7)** and **arm64**
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

#### 🚢 Container

Container images are available on
[ghcr.io](https://github.com/joshuar/go-hass-agent/pkgs/container/go-hass-agent).
Note that it is recommended to use an image tagged with the latest release
version over the latest container image, which might be unstable.

<!-- Usage -->
## :eyes: Usage

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
  --volume go-hass-agent:/home/ubuntu \
  ghcr.io/joshuar/go-hass-agent register \
  --server https://some.server:port \
  --token longlivedtoken
```

Adjust the `--server` and `--token` values as appropriate.

Once registered, run the agent with:

```shell
podman run --hostname go-hass-agent-container --name my-go-hass-agent \
  --network host \
  --volume go-hass-agent:/home/ubuntu \
  --volume /proc:/host/proc:ro --volume /sys:/host/sys:ro \
  --volume /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket:ro \
  --volume /run/user/1000/bus:/run/user/1000/bus:ro \
  ghcr.io/joshuar/go-hass-agent # any additional options
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

### Script Sensors

Go Hass Agent supports utilising scripts to create sensors. In this way, you can
extend the sensors presented to Home Assistant by the agent. Note that as the
agent is a “mobile app” in Home Assistant, any script sensors will be associated
with the Go Hass Agent device in Home Assistant.

Each script run by the agent can create one or more sensors and each script can
run on its own schedule, specified using a Cron-like syntax.

#### Requirements

- Scripts need to be put in
  `CONFIG_HOME/com.github.joshuar.go-hass-agent/scripts/` where `CONFIG_HOME`
  will be:
  - `~/.config` for Linux.
  - `~/Library/Application Support` for OSX.
  - `LocalAppData` for Windows.
- You can use symlinks, if supported by your Operating System.
- Script files need to be executable by the user running Go Hass Agent.
- Scripts need to run without any user interaction.
- Scripts need to output either valid JSON, YAML or TOML. See [Output
  Format](#output-format) for details.

#### Supported Scripting Languages

Any typical scripting language that can be invoked with a shebang can be used
for scripts. All scripts do not need to be written in the same language. So or
the typical shells can be used such as Bash, Sh, Zsh, Fish, Elvish. Scripting
languages such as Python, Perl and Ruby can also be used.

#### Output Format

All scripts should produce output that is either valid JSON, YAML or TOML.
Scripts do not need to use the same format; you can have one script that
produces JSON and another that produces TOML. All scripts will need to output
the following fields:

- A `schedule` field containing a [cron-formatted schedule](#schedule).
- A `sensors` field containing a list of sensors.

Sensors themselves need to be represented by the following fields:

- `sensor_name`: the *friendly* name of the sensor in Home Assistant (e.g., *My
  Script Sensor*).
- `sensor_icon`: a [Material Design
Icon](https://pictogrammers.github.io/@mdi/font/2.0.46/) representing the
current state. It can be changed dynamically based on the current state or
remain constant. Format is `mdi:icon_name`.
- `sensor_state`: the current value of the sensor. For numerical states, without
  the units. Otherwise, a *string* or *boolean* (for binary sensors).
  - **Note:** for a binary sensor, do not enclose the `true`/`false` in quotes.

The following optional fields can also be specified, which help control the
display in Home Assistant.

- `sensor_units`: the units for the state value.
- `sensor_type`: the *type* of sensor. If this is a binary sensor with a boolean
  value, set this to *“binary”*. Else, do not set this field.
- `sensor_device_class`: a Home Assistant [Device
Class](https://developers.home-assistant.io/docs/core/entity/sensor/#available-device-classes)
for the sensor, which will dictate how it will be displayed in Home Assistant.
There are many, pick an appropriate one (see
[`internal/hass/sensor/deviceClass.go`](../internal/hass/sensor/deviceClass.go)).
If setting `sensor_device_class`, it is likely required to set an appropriate
unit in `sensor_units` as well.
- `sensor_state_class`: the Home Assistant [State
  Class](https://developers.home-assistant.io/docs/core/entity/sensor/#available-state-classes).
  Either *measurement*, *total* or *total_increasing*.
- `sensor_attributes`: any additional attributes to be displayed with the
  sensor. **Note that the value is required to be valid JSON, regardless of the
  script output format.**

##### Examples

The following examples show a script that produces two sensors, in different
output formats.

###### JSON

JSON output can be either compressed:

```json
{"schedule":"@every 5s","sensors":[{"sensor_name": "random 1","sensor_icon": "mdi:dice-1","sensor_state":1},{"sensor_name": "random 2","sensor_icon": "mdi:dice-2","sensor_state_class":"measurement","sensor_state":6}]}
```

Or pretty-printed:

```json
{
  "schedule": "@every 5s",
  "sensors": [
    {
      "sensor_name": "random 1",
      "sensor_icon": "mdi:dice-1",
      "sensor_state": 2
    },
    {
      "sensor_name": "random 2",
      "sensor_icon": "mdi:dice-2",
      "sensor_state_class": "measurement",
      "sensor_state": 6
    }
  ]
}
```

###### YAML

```yaml
schedule: '@every 5s'
sensors:
    - sensor_name: random 1
      sensor_icon: mdi:dice-1
      sensor_state: 8
    - sensor_name: random 2
      sensor_icon: mdi:dice-2
      sensor_state_class: measurement
      sensor_state: 9
```

###### TOML

```toml
schedule = '@every 5s'

[[sensors]]
sensor_icon = 'mdi:dice-1'
sensor_name = 'random 1'
sensor_state = 3

[[sensors]]
sensor_icon = 'mdi:dice-2'
sensor_name = 'random 2'
sensor_state = 3
sensor_state_class = 'measurement'
```

For a binary sensor, the output should have `sensor_type` set to “binary” and
the `sensor_state` as `true` or `false` (without quotes). As an example in
compressed JSON format:

```json
{"schedule":"@every 10s","sensors":[{"sensor_name":"random 4","sensor_type":"binary","sensor_icon":"mdi:dice-3","sensor_state":false}]}
```

#### Schedule

The `schedule` field is used to specify the schedule or interval on which the
script will be run by the agent. Each script is run on its own schedule. All
sensors and their values should be returned each time the script is run. The
format is documented by the [cron Golang
package](https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format).
In most cases, it is presumed that the script needs to be run on some interval
of time. In that case, the easiest way to specify that is with the `@every
<duration>` as per the [example output](#examples) such as:

- `@every 5s`: every 5 seconds
- `@every 1h30m`: every 1 and a half hours.

Or a pre-defined schedule:

- `@hourly`.
- `@daily`.
- `@weekly`.
- `@monthly`.
- `@yearly`.

However, more cron formats are supported:

- `"30 * * * *"`: every hour on the half hour.
- `"30 3-6,20-23 * * *"`: in the range 3-6am, 8-11pm.
- `"CRON_TZ=Asia/Tokyo 30 04 * * *"`: at 04:30 Tokyo time every day.

> [!WARNING]
> Some schedules, while supported, might not make much sense.

#### Security Implications

Running scripts can be dangerous, especially if the script does not have robust
error-handling or whose origin is untrusted or unknown. Go Hass Agent makes no
attempt to do any analysis or sanitisation of script output, other than ensuring
the output is a [supported format](#output-format). As such, ensure you trust
and understand what the script does and all possible outputs that the script can
produce. Scripts are run by the agent and have the permissions of the user
running the agent. Script output is sent to your Home Assistant instance.

### Control via MQTT

> [!NOTE]
> Control via MQTT is not enabled by default.

If Home Assistant is connected to
[MQTT](https://www.home-assistant.io/integrations/mqtt/), you can also configure
Go Hass Agent to connect to MQTT, which will then allow you to run some commands
from Home Assistant to control the device running the agent.

#### Requirements

- For Linux:
  - Controls rely on distribution/system support for `systemd-logind` and a working
D-Bus connection.

#### Configuration

To configure the agent to connect to MQTT:

1. Right-click on the Go Hass Agent tray icon.
2. Select *Settings->App*.
3. Toggle ***Use MQTT*** and then enter the details for your MQTT server (not
   your Home Assistant server).
4. Click ***Save***.
5. Restart Go Hass Agent.

After the above steps, Go Hass Agent will appear as a device under the MQTT
integration in your Home Assistant.

[![Open your Home Assistant instance and show the MQTT integration.](https://my.home-assistant.io/badges/integration.svg)](https://my.home-assistant.io/redirect/integration/?domain=mqtt)

> [!NOTE]
> Go Hass Agent will appear in two places in your Home Assistant.
> Firstly, under the Mobile App integration, which will show all the *sensors*
> that Go Hass Agent is reporting. Secondly, under the MQTT integration, which
> will show the *controls* for Go Hass Agent. Unfortunately, due to limitations
> with the Home Assistant architecture, these cannot be combined in a single
> place.

#### Available Controls

The following table shows the controls that are available.  You can add these
controls to dashboards in Home Assistant or use them in automations with a
service call.

| Control | What it does |
|--------|------------------|
| Lock Screen | Locks the session for the user running Go Hass Agent |
| Unlock Screen | Unlocks the session for the user running Go Hass Agent |
| Lock Screensaver | Lock the “screensaver” of the session for the user running Go Hass Agent |
| Suspend | Will (instantly) suspend (the system state is saved to RAM and the CPU is turned off) the device running Go Hass Agent |
| Hibernate | Will (instantly) hibernate (the system state is saved to disk and the machine is powered down) the device running Go Hass Agent |
| Power Off | Will (instantly) power off the device running Go Hass Agent |
| Reboot | Will (instantly) reboot the device running Go Hass Agent |
| Volume Control | Adjust the volume on the default audio output device |
| Volume Mute | Mute/Unmute the default audio output device |

#### Arbitrary D-BUS commands

The agent will subscribe to the MQTT topic `gohassagent/dbus` on the configured
MQTT broker and listens for JSON messages of the below format, which will be
accordingly dispatched to the systems
[D-Bus](https://www.freedesktop.org/wiki/Software/dbus/).

```json
{
  "bus": "session",
  "path": "/org/cinnamon/ScreenSaver",
  "method": "org.cinnamon.ScreenSaver.Lock",
  "destination": "org.cinnamon.ScreenSaver",
  "args": [
    ""
  ]
}
```

This can be used to trigger arbitrary d-bus commands on the system where the
agent runs on, by using any MQTT client such as the Home Assistant
[`mqtt.publish`](https://www.home-assistant.io/integrations/mqtt/#service-mqttpublish)
service.

#### Security Implications

There is a significant discrepancy in permissions between the device running Go
Hass Agent and Home Assistant.

Go Hass Agent runs under a user account on a device. So the above controls will
only work where that user has permissions to run the underlying actions on that
device. Home Assistant does not currently offer any fine-grained access control
for controls like the above. So any Home Assistant user will be able to run any
of the controls. This means that a Home Assistant user not associated with the
device user running the agent can use the exposed controls to issue potentially
disruptive actions on a device that another user is accessing.

<!-- Roadmap -->
## :compass: Roadmap

Check out [what I'm working
on](https://github.com/joshuar/go-hass-agent/discussions/150) for future
releases.

<!-- Contributing -->
## :wave: Contributing

- Found an issue? Please [report
  it](https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D)!
- Have a suggestion for a feature? Want a particular sensor/measurement added?
  Submit a [feature
  request](https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=feature_request.md&title=)!
- Want to help develop Go Hass Agent? See the [contributing
  guidelines](CONTRIBUTING.md).

<!-- Code of Conduct -->
### :scroll: Code of Conduct

Please read the [Code of Conduct](https://github.com/joshuar/go-hass-agent/blob/master/CODE_OF_CONDUCT.md)

<!-- FAQ -->
## :grey_question: FAQ

- _Can I change the units of the sensor?_
  - Yes! In the [customisation
options](https://www.home-assistant.io/docs/configuration/customizing-devices/)
for a sensor/entity, you can change the _unit of measurement_ (and _display
precision_ if desired). This is useful for sensors whose native unit is not very
human-friendly. For example the memory sensors report values in bytes (B),
whereas you may wish to change the unit of measurement to gigabytes (GB).

- _Can I disable some sensors?_
  - The agent itself does not currently support disabling individual sensors.
   However, you can disable the corresponding sensor entity in Home Assistant,
   and the agent will stop sending updates for it.
  - To disable a sensor entity, In the [customisation
options](https://www.home-assistant.io/docs/configuration/customizing-devices/)
for a sensor/entity, toggle the *Enabled* switch. The agent will automatically
detect the disabled state and send/not send updates as appropriate.
  - Note that while the agent will stop sending updates for a disabled sensor,
it will not stop gathering the raw data for the sensor.

- _The GUI windows are too small/too big. How can I change the size?_
  - See [Scaling](https://developer.fyne.io/architecture/scaling) in the Fyne
documentation. In the tray icon menu, select _Settings_ to open the Fyne
settings app which can adjust the scaling for the app windows.

- _What is the resource (CPU, memory) usage of the agent?_
  - Very little in most cases. On Linux, the agent with all sensors working, should
consume well less than 50 MB of memory with very little CPU usage. Further
memory savings can be achieved by running the agent in “headless” mode with the
`--terminal` command-line option. This should put the memory usage below 25 MB.
  - On Linux, many sensors rely on D-Bus signals for publishing their data, so
CPU usage may be affected by the “business” of the bus. For sensors that are
polled on an interval, the agent makes use of some jitter in the polling
intervals to avoid a “thundering herd” problem.

- _I've updated the agent and now some sensors have been renamed. I now have a
  bunch of sensors/entities in Home Assistant I want to remove. What can I do?_
  - Unfortunately, sometimes the sensor naming scheme for some sensors created
by the agent needs to change. There is unfortunately, no way for the agent to
rename existing sensors in Home Assistant, so you end up with both the old and
new sensors showing, and only the new sensors updating.
  - You can remove the old sensors manually, under Developer Tools→Statistics in
Home Assistant, for example. The list should contain sensors that are no longer
“provided” by the agent. Or you can wait until they age out of the Home
Assistant long-term statistics database automatically.

- _Can I reset the agent (start from new)?_
  - Yes. You can reset the agent so that it will re-register with Home Assistant
and act as a new device. To do this:
    1. Shut down the agent if it is running.
    2. In Home Assistant, navigate to **Settings→Devices & Services** and click
       on the **Mobile App** integration.
    3. Locate the agent entry in the list of mobile devices, click the context menu
       (three vertical dots), and choose **_Delete_**.
    4. From a terminal, run the agent with the command: `go-hass-agent register
       --force` (add `--terminal --server someserver --token sometoken` for
       non-graphical registration).
    5. The agent will go through the initial registration steps. It should report
       that registration was successful.
    6. Restart the agent.

- _I want to run the agent on a server, as a service, without a GUI. Can I do
  this?_
  - Yes. The packages install a systemd service file that can be enabled and
used to run the agent as a service.
  - You will still need to register the agent manually before starting as a service.
See the command for registration in the [README](../README.md#running-headless).
  - You will also need to ensure your user has “lingering” enabled.  Run `loginctl
list-users` and check that your user has **LINGER** set to “yes”. If not, run
`loginctl enable-linger`.
  - Once you have registered the agent and enabled lingering for your user. Enable
the service and start it with the command: `systemctl --user enable
go-hass-agent && systemctl --user start go-hass-agent`.
  - You can check the status with `systemctl --user status go-hass-agent`. The agent
should start with every boot.
  - For other init systems, consult their documentation on how to enable and run
user services.

<!-- Acknowledgments -->
## :gem: Acknowledgements

- [Home Assistant](https://home-assistant.io).
- This [Awesome README Template](https://github.com/Louis3797/awesome-readme-template).

<!-- License -->
## :warning: License

[MIT](LICENSE)
