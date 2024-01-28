<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Control via MQTT

If Home Assistant is connected to
[MQTT](https://www.home-assistant.io/integrations/mqtt/), you can also configure
Go Hass Agent to connect to MQTT, which will then allow you to run some commands
from Home Assistant to control the device running the agent.

**Control via MQTT is not enabled by default.**

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

## Available Controls

The following table shows the controls that are available.  You can add these
controls to dashboards in Home Assistant or use them in automations with a
service call.

| Control | What it does |
|--------|------------------|
| Lock Screen | Locks the session for the user running Go Hass Agent |
| UnLock Screen | Unlocks the session for the user running Go Hass Agent |
| Power Off | Will power off the device running Go Hass Agent |
| Reboot | Will reboot the device running Go Hass Agent |

## Security

There is a significant discrepancy in permissions between the device running Go Hass Agent and Home Assistant.

Go Hass Agent runs under a user account on a device. So the above controls will only work where that user has permissions to run the underlying actions on that device. Home Assistant does not currently offer any fine-grained access control for controls like the above. So any Home Assistant user will be able to run any of the controls. This means that a Home Assistant user not associated with the device user running the agent can use the exposed controls to issue potentially disruptive actions on a device that another user is accessing.

## Implementation Details

### Linux

Controls rely on distribution/system support for `systemd-logind` and a working D-Bus connection.
