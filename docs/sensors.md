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

| Sensor | What it measures | Source | Extra Attributes | Update Frequency |
|--------|------------------|--------|-------------------|-------------------|
| Agent Version | The version of Go Hass Agent |  | | On agent start. |
| Active App | Currently active (focused) application | D-Bus | | When app changes. |
| Running Apps | Count of all running applications | D-Bus | The application names | When running apps count changes. | 
| Accent Color  | The hex code representing the accent color of the desktop environment in use. | D-Bus | | When accent color changes. | 
| Theme Type  | Whether a dark or light desktop theme is detected. | D-Bus | | When desktop theme changes. | 
| Battery Type | The type of battery (e.g., UPS, line power) | D-Bus | | On battery addeded/removed. |
| Battery Temp | The current battery temperature | D-Bus | | When temp changes. |
| Battery Power | The battery current power draw | D-Bus | Voltage, Energy consumption, where reported | When voltage changes. |
| Battery Level/Percentage | The current battery capacity | D-Bus | | When level changes. |
| Battery State | The current battery state (e.g., charging/discharging) | D-Bus | | When state changes. |
| Memory Total | Total memory on the system | ProcFS | | ~Every minute |
| Memory Available | Memory available/free | ProcFS | | ~Every minute |
| Memory Used | Memory used | ProcFS | | ~Every minute |
| Memory Usage | Total memory usage % | ProcFS | | ~Every minute |
| Swap Total | Total swap on the system | ProcFS | | ~Every minute |
| Swap Available | Swap available/free | ProcFS | | ~Every minute |
| Swap Used | Swap used | ProcFS | | ~Every minute |
| Swap Usage | Swap memory usage % | ProcFS | | ~Every minute |
| Per Mountpoint Usage | % usage of mount point. | ProcFS |  Filesystem type, bytes/inode total/free/used. | ~Every minute. |
| Device total read/writes and rates | Count of read/writes, Rate (in KB/s) of reads/writes, to the device. | ProcFS | | ~Every 5 seconds. |
| Connection State (per-connection) | The current state of each network connection | D-Bus | Connection type (e.g., wired/wireless/VPN), IP addresses | When connections change. |
| Wi-Fi SSID[^1] | The SSID of the Wi-Fi network | D-Bus | | When SSID changes. |
| Wi-Fi Frequency[^1] | The frequency band of the Wi-Fi network | D-Bus | | When frequency changes. | 
| Wi-Fi Speed[^1] | The network speed of the Wi-Fi network | D-Bus | | When speed changes. |
| Wi-Fi Strength[^1] | The strength of the signal of the Wi-Fi network | D-Bus | | When strength changes. |
| Wi-Fi BSSID[^1] | The BSSID of the Wi-Fi network | D-Bus | | When BSSID changes. |
| Bytes Received | Total bytes received | ProcFS | Packet count, drops, errors | ~Every 5 seconds. |
| Bytes Sent | Total bytes sent | ProcFS | Packet count, drops, errors | ~Every 5 seconds. |
| Bytes Received Rate | Current received transfer rate  | ProcFS | | ~Every 5 seconds. |
| Bytes Sent Rate | Current sent transfer rate | ProcFS | | ~Every 5 seconds. |
| Load Average 1min | 1min load average | ProcFS |  | ~Every 1 minute. |
| Load Average 5min | 5min load average | ProcFS |  | ~Every 1 minute. |
| Load Average 15min | 15min load average | ProcFS |  | ~Every 1 minute. |
| CPU Usage | Total CPU Usage % | ProcFS | | ~Every 10 seconds. |
| Power Profile | The current power profile as set by the power-profiles-daemon | D-Bus | | When profile changes. |
| Boot Time | Date/Time of last system boot | ProcFS |  | ~Every 15 minutes. |
| Uptime | System uptime | ProcFS | | ~Every 15 minutes. |
| Kernel Version | Version of the currently running kernel | ProcFS | | On agent start. |
| Distribution Name | Name of the running distribution (e.g., Fedora, Ubuntu) | ProcFS | | On agent start. |
| Distribution Version | Version of the running distribution | ProcFS | | On agent start. |
| Current Users | Count of active users on the system | D-Bus | List of usernames | When user count changes. |
| Screen Lock State | Current state of screen lock | D-Bus | | When screen lock changes. |
| Power State | Power state of device (e.g., suspended, powered on/off) | D-Bus | | When power state changes. |
| Problems | Count of any problems logged to the ABRT daemon | D-Bus |  Problem details | ~Every 15 minutes |
| Device/Component Sensors(s) | Any reported hardware sensors (temp, fan speed, voltage, etc.) from each device/component, as extracted from the `/sys/class/hwmon` file system. | SysFS |  | ~Every 1 minute. |

[^1]: Only updated when currently connected to a Wi-Fi network.

## Scripts (All Platforms)

All platforms can also utilise scripts to create custom sensors. See [scripts](scripts.md).
