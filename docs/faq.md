<!--
 Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Frequently Asked Questions

## Q: Can I change the units of the sensor?

Yes! In the [customisation
options](https://www.home-assistant.io/docs/configuration/customizing-devices/)
for a sensor/entity, you can change the _unit of measurement_ (and _display
precision_ if desired). This is useful for sensors whose native unit is not very
human-friendly. For example the memory sensors report values in bytes (B), whereas
you may wish to change the unit of measurement to gigabytes (GB).

## Q: Can I disable some sensors?

The agent itself does not currently support disabling individual sensors.
However, you can disable the corresponding sensor entity in Home Assistant, and
the agent will stop sending updates for it. 

To disable a sensor entity, In the [customisation
options](https://www.home-assistant.io/docs/configuration/customizing-devices/)
for a sensor/entity, toggle the *Enabled* switch. The agent will automatically
detect the disabled state and send/not send updates as appropriate.

Note that while the agent will stop sending updates for a disabled sensor, it
will not stop gathering the raw data for the sensor.

## Q: The GUI windows are too small/too big. How can I change the size?

See [Scaling](https://developer.fyne.io/architecture/scaling) in the Fyne
documentation. In the tray icon menu, select _Settings_ to open the Fyne
settings app which can adjust the scaling for the app windows.

## Q: What is the resource (CPU, memory) usage of the agent?

Very little in most cases. On Linux, the agent with all sensors working, should
consume well less than 50 MB of memory with very little CPU usage. Further
memory savings can be achieved by running the agent in “headless” mode with the
`--terminal` command-line option. This should put the memory usage below 25 MB.

On Linux, many sensors rely on D-Bus signals for publishing their data, so CPU
usage may be affected by the “business” of the bus. For sensors that are polled
on an interval, the agent makes use of some jitter in the polling intervals to
avoid a “thundering herd” problem.

## Q: I've updated the agent and now some sensors have been renamed. I now have a bunch of sensors/entities in Home Assistant I want to remove. What can I do?

Unfortunately, sometimes the sensor naming scheme for some sensors created by
the agent needs to change. There is unfortunately, no way for the agent to
rename existing sensors in Home Assistant, so you end up with both the old and
new sensors showing, and only the new sensors updating.

You can remove the old sensors manually, under Developer Tools→Statistics in
Home Assistant, for example. The list should contain sensors that are no longer
“provided” by the agent. Or you can wait until they age out of the Home
Assistant long-term statistics database automatically.

## Q: Can I reset the agent (start from new)?

Yes. You can reset the agent so that it will re-register with Home Assistant and
act as a new device. To do this:

1. Shut down the agent if it is running.
2. In Home Assistant, navigate to **Settings→Devices & Services** and click on the
   **Mobile App** integration.
3. Locate the agent entry in the list of mobile devices, click the context menu
   (three vertical dots), and choose ***Delete***.
4. From a terminal, run the agent with the following arguments:

```shell
# add --terminal --server someserver --token sometoken for non graphical registration
go-hass-agent register --force 
```

5. The agent will go through the initial registration steps. It should report
   that registration was successful.
6. Restart the agent.

## Q: I want to run the agent on a server, as a service, without a GUI. Can I do this?

Yes. The packages install a systemd service file that can be enabled and used to
run the agent as a service. 

You will still need to register the agent manually before starting as a service.
See the command for registration in the [README](../README.md#running-headless).

You will also need to ensure your user has “lingering” enabled.  Run `loginctl
list-users` and check that your user has **LINGER** set to “yes”. If not, run
`loginctl enable-linger`.

Once you have registered the agent and enabled lingering for your user. Enable
the service and start it:

```shell
systemctl --user enable go-hass-agent
systemctl --user start go-hass-agent
```

You can check the status with `systemctl --user status go-hass-agent`. The agent
should start with every boot.

For other init systems, consult their documentation on how to enable and run
user services.



