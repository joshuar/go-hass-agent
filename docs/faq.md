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

You can remove the old sensors manually, under Developer Tools->Statistics in
Home Assistant, for example. The list should contain sensors that are no longer
"provided" by the agent. Or you can wait until they age out of the Home
Assistant long-term statistics database automatically.