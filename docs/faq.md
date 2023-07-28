<!--
 Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Frequently Asked Questions

#### Q: Can I change the units of the sensor?

Yes! In the [customisation
options](https://www.home-assistant.io/docs/configuration/customizing-devices/)
for a sensor/entity, you can change the _unit of measurement_ (and _display
precision_ if desired). This is useful for sensors whose native unit is not very
human-friendly. For example the memory sensors report values in bytes (B), whereas
you may wish to change the unit of measurement to gigabytes (GB).

#### Q: The GUI windows are too small/too big. How can I change the size?

See [Scaling](https://developer.fyne.io/architecture/scaling) in the Fyne
documentation. In the tray icon menu, select _Settings_ to open the Fyne
settings app which can adjust the scaling for the app windows.
