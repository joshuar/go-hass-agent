<!--
 Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Frequently Asked Questions

## Data/Sensors Reported

#### Q: Could this implement reporting system metrics like CPU, RAM and Disk metrics?

There are a lot of existing Home Assistant integrations that publish such
metrics. As such, this app probably won't implement sending such metrics.

#### Q: Can I change the units of the sensor?

Yes! In the [customisation
options](https://www.home-assistant.io/docs/configuration/customizing-devices/)
for a sensor/entity, you can change the *unit of measurement* (and *display
precision* if desired). This is useful for sensors whose native unit is not very
human-friendly. For example the memory sensors report values in bytes (B), whereas
you may wish to change the unit of measurement to gigabytes (GB).

## GUI

#### Q: The GUI windows are too small/too big. How can I change the size?

See [Scaling](https://developer.fyne.io/architecture/scaling) in the Fyne
documentation. Use the `fyne_settings` app to adjust the size of the displayed
windows and their widgets.
