<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# By Operating System

## Linux

> [!NOTE]
> The following list shows all **potential** sensors the agent can
> report. In some cases, the **actual** sensors reported will be less due to
> lack of support or missing hardware.

| Sensor | What it measures | Source | Extra Attributes |
|--------|------------------|--------|-------------------|
| Active App | Currently active (focused) application | D-Bus | |
| Running Apps | Count of all running applications | D-Bus | The application names |
| Battery Type | The type of battery (e.g., UPS, line power) | D-Bus | |
| Battery Temp | The current battery temperature | D-Bus | |
| Battery Power | The battery current power draw | D-Bus | Voltage, Energy consumption, where reported |
| Battery Level/Percentage | The current battery capacity | D-Bus | |
| Battery State | The current battery state (e.g., charging/discharging) | D-Bus | |
| Memory Total | Total memory on the system | ProcFS | |
| Memory Available | Memory available/free | ProcFS | |
| Memory Used | Memory used | ProcFS | |
| Memory Usage | Total memory usage % | ProcFS | |
| Swap Total | Total swap on the system | ProcFS | |
| Swap Available | Swap available/free | ProcFS | |
| Swap Used | Swap used | ProcFS | |
| Swap Usage | Swap memory usage % | ProcFS | |
| Per Mountpoint Usage | % usage of mount point | ProcFS |  Filesystem type, bytes/inode total/free/used |
| Connection State (per-connection) | The current state of each network connection | D-Bus | Connection type (e.g., wired/wireless/VPN), IP addresses |
| Wi-Fi SSID[^1] | The SSID of the Wi-Fi network | D-Bus | |
| Wi-Fi Frequency[^1] | The frequency band of the Wi-Fi network | D-Bus | |
| Wi-Fi Speed[^1] | The network speed of the Wi-Fi network | D-Bus | |
| Wi-Fi Strength[^1] | The strength of the signal of the Wi-Fi network | D-Bus | |
| Wi-Fi BSSID[^1] | The BSSID of the Wi-Fi network | D-Bus | |
| Bytes Received | Total bytes received | ProcFS | Packet count, drops, errors |
| Bytes Sent | Total bytes sent | ProcFS | Packet count, drops, errors |
| Bytes Received Rate | Current received transfer rate  | ProcFS | |
| Bytes Sent Rate | Current sent transfer rate | ProcFS | |
| Load Average 1min | 1min load average | ProcFS |  |
| Load Average 5min | 5min load average | ProcFS |  |
| Load Average 15min | 15min load average | ProcFS |  |
| CPU Usage | Total CPU Usage % | ProcFS | |
| Power Profile | The current power profile as set by the power-profiles-daemon | D-Bus | |
| Boot Time | Date/Time of last system boot | ProcFS |  |
| Uptime | System uptime | ProcFS | |
| Kernel Version | Version of the currently running kernel | ProcFS | |
| Distribution Name | Name of the running distribution (e.g., Fedora, Ubuntu) | ProcFS | |
| Distribution Version | Version of the running distribution | ProcFS | |
| Current Users | Count of active users on the system | D-Bus | List of usernames |
| Screen Lock State | Current state of screen lock | D-Bus | |
| Power State | Power state of device (e.g., suspended, powered on/off) | D-Bus | |
| Problems | Count of any problems logged to the ABRT daemon | D-Bus |  Problem details |
| Device/Component Sensors(s) | Any reported hardware sensors (temp, fan speed, voltage, etc.) from each device/component, as extracted from the `/sys/class/hwmon` file system. | SysFS |  |

[^1]: Only updated when currently connected to a Wi-Fi network.

## Scripts (All Platforms)

All platforms can also utilise scripts to create custom sensors. See [scripts](scripts.md).
