# Changelog

## [14.9.0](https://github.com/joshuar/go-hass-agent/compare/v14.8.0...v14.9.0) (2026-02-21)


### Features

* **linux/disk:** :sparkles: add attributes to disk usage sensors for usage/free/totals in bytes ([6c0bf7f](https://github.com/joshuar/go-hass-agent/commit/6c0bf7f0dea052b748246570ea650824dfc750db))


### Bug Fixes

* :bug: improvement registration logic ([884fbb4](https://github.com/joshuar/go-hass-agent/commit/884fbb4ee56f7f45a712d5211b26e9296f401a34))
* **linux/system:** :bug: open input devices read-only and non-blocking ([5e54843](https://github.com/joshuar/go-hass-agent/commit/5e54843e7b4bddd06b7d3ec7b68157096d105520))

## [14.8.0](https://github.com/joshuar/go-hass-agent/compare/v14.7.0...v14.8.0) (2026-02-15)


### Features

* **linux/system:** :sparkles: improved user activity worker ([afdc6c0](https://github.com/joshuar/go-hass-agent/commit/afdc6c0243d625e4379ea2d16bbfd8ab2956bfaf))


### Bug Fixes

* **linux/system:** :bug: fix clean-up on shutdown for user activity worker ([2ead3a1](https://github.com/joshuar/go-hass-agent/commit/2ead3a12522e716ecf0f504218f0db3c7688846a))

## [14.7.0](https://github.com/joshuar/go-hass-agent/compare/v14.6.0...v14.7.0) (2026-02-08)


### Features

* **linux:** :sparkles: try to detect desktop portal implementation and provide option to override in preferences if necessary ([bd3c675](https://github.com/joshuar/go-hass-agent/commit/bd3c675bc2137bf97ffc51e5b41d1d672ffb6f07))


### Bug Fixes

* **cli:** :bug: fix display of registration status ([a5c4fe5](https://github.com/joshuar/go-hass-agent/commit/a5c4fe52852e8bc5af734553d1f7889464da3648))
* **linux/disk:** :bug: fix detection of ignored mounts ([df2e8e8](https://github.com/joshuar/go-hass-agent/commit/df2e8e80703ad078e00a26fdefc05236cd120767))
* **server:** :bug: correct address shown for where registration url ([3868d60](https://github.com/joshuar/go-hass-agent/commit/3868d605a67c03ca27ea0641c25fe945232df0c5))

## [14.6.0](https://github.com/joshuar/go-hass-agent/compare/v14.5.0...v14.6.0) (2026-01-31)


### Features

* **web:** :sparkles: add simple landing/home page with table of sensor values ([e961ff3](https://github.com/joshuar/go-hass-agent/commit/e961ff38390380f6ac9ed9035972b5cfd3f08571))


### Bug Fixes

* **hass/api:** :bug: don't send parts of request with nil/empty values ([6da589c](https://github.com/joshuar/go-hass-agent/commit/6da589c10cf2f0a5f418aac332df9413b4cd9977))

## [14.5.0](https://github.com/joshuar/go-hass-agent/compare/v14.4.0...v14.5.0) (2026-01-31)


### Features

* **cli:** :sparkles: add a cli command to list the entities in the local registry and their state ([6bb1cd2](https://github.com/joshuar/go-hass-agent/commit/6bb1cd20f8bf1853594b3c6f105b347f276ae276))


### Bug Fixes

* **config:** :bug: config fixes ([6792ac5](https://github.com/joshuar/go-hass-agent/commit/6792ac50be333efccd00ea0c19aad0428fa3213f))
* **config:** :bug: fix permissions on `~/.config/go-hass-agent` directory on creation ([1fef209](https://github.com/joshuar/go-hass-agent/commit/1fef209da1ef25f9cf81c4ba5b9daa45fc7532a5))
* **linux/disk:** :bug: also exclude exact matches for ignored mount points in disk usage sensor ([b589a19](https://github.com/joshuar/go-hass-agent/commit/b589a19f56b46a2260632812f9e76f3e69e2efe1))


### Performance Improvements

* **hass:** :sparkles: support being able fetch hass client in multiple places, but only init it once ([6fad90b](https://github.com/joshuar/go-hass-agent/commit/6fad90b942eab56fef39d2028da811dbe3996d34))

## [14.4.0](https://github.com/joshuar/go-hass-agent/compare/v14.3.1...v14.4.0) (2026-01-26)


### Features

* add an AppStream MetaInfo file ([6ef67a8](https://github.com/joshuar/go-hass-agent/commit/6ef67a8fb651e29bfb6a8997accffe790db000b8))
* add an AppStream MetaInfo file ([98cb7f2](https://github.com/joshuar/go-hass-agent/commit/98cb7f27cf61877608172cac04610a08d0b77eaf)), closes [#703](https://github.com/joshuar/go-hass-agent/issues/703)
* **agent:** :sparkles: on first run, either open a browser window to the registration page or emit a log message asking the user to do so ([2334a3a](https://github.com/joshuar/go-hass-agent/commit/2334a3a5402412809f266b5c8175d0f6a23191d1))


### Bug Fixes

* **linux/hwmon:** :bug: fix when warning about device model is shown ([8cb0669](https://github.com/joshuar/go-hass-agent/commit/8cb0669a24c7abb46e8e25aace97a43c82cceafb))
* **linux/net:** :bug: would it be silly to assume strings.HasPrefix would return true when the string equals the prefix? yes, yes it would ([3b6fbb3](https://github.com/joshuar/go-hass-agent/commit/3b6fbb35556423b41ec3f87a03305d5ec48ad175))
* **linux/power:** :bug: add a check to ensure system has required brightness controls when creating brightness worker ([d5c8f6f](https://github.com/joshuar/go-hass-agent/commit/d5c8f6ffae0700cde2223abc8559545c737161f0))
* **linux/power:** :bug: update D-Bus location for brightness control as per changes in Gnome 49?? ([bdc8d55](https://github.com/joshuar/go-hass-agent/commit/bdc8d5519c2ec2d4947e380a70f10916d4165625))


### Reverts

* **linux/power:** :rewind: give up using D-Bus on Gnome for backlight control, just try ddcutil ([4842317](https://github.com/joshuar/go-hass-agent/commit/4842317ece2140419ef998a0173cfa74b2dc93f5))

## [14.3.1](https://github.com/joshuar/go-hass-agent/compare/v14.3.0...v14.3.1) (2025-12-26)


### Bug Fixes

* :bug: don't return nil worker when creating new workers to avoid nil pointer panic ([d8ebd17](https://github.com/joshuar/go-hass-agent/commit/d8ebd17567f13211e503690161dcd39ad3e25ae1))

## [14.3.0](https://github.com/joshuar/go-hass-agent/compare/v14.2.3...v14.3.0) (2025-12-20)


### Features

* **linux/power:** :sparkles: implement monitor backlight brightness control/state ([86d42fc](https://github.com/joshuar/go-hass-agent/commit/86d42fcf967ff7ade4852b3c19a00291f1d96b7f))


### Bug Fixes

* **linux/power:** :bug: fix nil pointer when power profile sensor settings can't be read ([02585b2](https://github.com/joshuar/go-hass-agent/commit/02585b28603572b8d225e9ce240d0eed3dd16efd))


### Performance Improvements

* **linux/hwmon:** :zap: use bytes.Buffer rather than direct []byte for hwmon file pool ([77085a9](https://github.com/joshuar/go-hass-agent/commit/77085a9a2812ef53e19b376362306a8d0a209dff))

## [14.2.3](https://github.com/joshuar/go-hass-agent/compare/v14.2.2...v14.2.3) (2025-12-19)


### Bug Fixes

* **config:** mark embedded config structs as ",squash" thanks [@cmnrd](https://github.com/cmnrd)! ([dce8c09](https://github.com/joshuar/go-hass-agent/commit/dce8c09bfbf8f17e1c433b2577ad4f0920aeb865))
* **config:** mark embedded config structs as "squash" ([f780f3e](https://github.com/joshuar/go-hass-agent/commit/f780f3eeea89fab74f4412212c096c9cf00d55d8)), closes [#624](https://github.com/joshuar/go-hass-agent/issues/624)


### Performance Improvements

* **linux/hwmon:** :zap: more optimisations ([edcf082](https://github.com/joshuar/go-hass-agent/commit/edcf082c59c1fd6f0f72d3d3956564a725cd76e1))
* **pkg/hwmon:** :zap: hwmon fetching improvements ([59f494e](https://github.com/joshuar/go-hass-agent/commit/59f494e9fd93ba697dfed3d676c3a430972fa64a))

## [14.2.2](https://github.com/joshuar/go-hass-agent/compare/v14.2.1...v14.2.2) (2025-12-13)


### Bug Fixes

* **linux/net:** :bug: reworked networkmanager connection sensor ([91ef151](https://github.com/joshuar/go-hass-agent/commit/91ef1511139f5994563acbfadadcf5aabfc06937))
* **linux/net:** :recycle: rename pkg ([f7ca26f](https://github.com/joshuar/go-hass-agent/commit/f7ca26f975a29f48d3d6abded0ba37e4f76e1518))

## [14.2.1](https://github.com/joshuar/go-hass-agent/compare/v14.2.0...v14.2.1) (2025-12-07)


### Bug Fixes

* **linux/power:** don't pass empty args to screen control commands ([ba9a03a](https://github.com/joshuar/go-hass-agent/commit/ba9a03aaa9e55691d07fb4bfdc8aa5f357bc8c7b))

## [14.2.0](https://github.com/joshuar/go-hass-agent/compare/v14.1.1...v14.2.0) (2025-11-28)


### Features

* **server:** :sparkles: add /health-check endpoint for checking if agent is running/responding ([0435ca2](https://github.com/joshuar/go-hass-agent/commit/0435ca2cbcd60404667c3617c99be70203be3548))
* **server:** :sparkles: support configuring web server options (https cert/key, hostname, port) on command-line ([0dfcaa5](https://github.com/joshuar/go-hass-agent/commit/0dfcaa564f74815fac4c01639b7592486d6e4f74))

## [14.1.1](https://github.com/joshuar/go-hass-agent/compare/v14.1.0...v14.1.1) (2025-11-07)


### Bug Fixes

* :bug: fix mqtt preferences issues ([5ffc763](https://github.com/joshuar/go-hass-agent/commit/5ffc76340ce9f5881253d4913e8b407d9cc040a1))
* **linux/power:** :bug: add args to screen lock D-Bus calls (needed for screen lock on Cinnamon) ([effaeec](https://github.com/joshuar/go-hass-agent/commit/effaeec5494c4af72a83195c7cd2b8737d11ed7c))

## [14.1.0](https://github.com/joshuar/go-hass-agent/compare/v14.0.3...v14.1.0) (2025-10-12)


### Features

* **container:** :building_construction: don't log to a file when running in a container (still log to stderr) ([fdf8c1d](https://github.com/joshuar/go-hass-agent/commit/fdf8c1ddd85e17d88c14e19b200c6ef4cbef5a91))
* **linux/disk:** :sparkles: support a user preference for ignoring mountpoints for disk usage sensor ([17d4490](https://github.com/joshuar/go-hass-agent/commit/17d449074a218a259a4aa43dded9b10bb717b86f))


### Bug Fixes

* **linux/power:** :bug: fix spelling mistake in unlocking session command for KDE/Gnome (╯°□°）╯︵ ┻━┻ ([38d2dd6](https://github.com/joshuar/go-hass-agent/commit/38d2dd6d30a9424c6c24183c06b7e2d1d2c19f1d))

## [14.0.3](https://github.com/joshuar/go-hass-agent/compare/v14.0.2...v14.0.3) (2025-10-05)


### Bug Fixes

* **agent:** :sparkles: re-add notification support (when running on a desktop) ([416901f](https://github.com/joshuar/go-hass-agent/commit/416901f31c4e45d7c513931235c83e2c776e2ab0))
* **linux/cpu:** :bug: fix preferences prefix for cpu frequency sensor preferences ([95c6184](https://github.com/joshuar/go-hass-agent/commit/95c61847ecb2ab30f680f0d58ea345dd4c5b9933))
* **linux/net:** :bug: fix filtering of devices for link sensors ([4585133](https://github.com/joshuar/go-hass-agent/commit/45851338fd97d5ef5ad5bccd400cf940397c6f31))

## [14.0.2](https://github.com/joshuar/go-hass-agent/compare/v14.0.1...v14.0.2) (2025-09-30)


### Bug Fixes

* **linux/cpu:** :bug: correctly store previously cpu time measurement for usage calculation ([00d6c61](https://github.com/joshuar/go-hass-agent/commit/00d6c61086689eb76104672156ec5c9ecfe1f638))

## [14.0.1](https://github.com/joshuar/go-hass-agent/compare/v14.0.0...v14.0.1) (2025-09-29)


### Bug Fixes

* **linux/cpu:** :bug: correct CPU usage values ([12b2da9](https://github.com/joshuar/go-hass-agent/commit/12b2da91be45ae9816d9afd0da514dd63e94deaa))

## [14.0.0](https://github.com/joshuar/go-hass-agent/compare/v13.3.3...v14.0.0) (2025-09-28)


### Features

* :bento: add additional frontend assets ([32bdffd](https://github.com/joshuar/go-hass-agent/commit/32bdffd0a93a7451444c2358579d53f20fb23d1d))
* :sparkles: support environment variables for setting log level and logging to file ([c936223](https://github.com/joshuar/go-hass-agent/commit/c936223994924f9fe26afe4603f13c97273d4769))
* :sparkles: web-based registration start ([c675b71](https://github.com/joshuar/go-hass-agent/commit/c675b71aebf3d06f92909af0e14b50a7a77fe4eb))
* **agent:** :sparkles: generate and save a device ID as necessary ([27e4b4c](https://github.com/joshuar/go-hass-agent/commit/27e4b4cc01e9fbcf500e579369bfec8f3ee46a19))
* **cli:** :sparkles: re-add config command ([b98915f](https://github.com/joshuar/go-hass-agent/commit/b98915f66df46bfa0b2fcbab2fd793757c02e9cf))
* **cli:** :sparkles: re-add register command ([1b74186](https://github.com/joshuar/go-hass-agent/commit/1b741862dcad94302d07154ab05614dce146c94d))
* **dbusx:** :sparkles: add a function to get all methods on a bus ([aed572d](https://github.com/joshuar/go-hass-agent/commit/aed572d4ebc341659f2f140113ad0d0090649c0c))
* **device:** :sparkles: re-add device info methods ([7dec66f](https://github.com/joshuar/go-hass-agent/commit/7dec66f3461eb283d07c05fa091c3eadb7ef0fd5))
* **server:** :sparkles: implement preferences page ([2b7ee34](https://github.com/joshuar/go-hass-agent/commit/2b7ee34fe75f3fc926471472c71daf08a3430435))
* **server:** :sparkles: server improvements ([f5cf330](https://github.com/joshuar/go-hass-agent/commit/f5cf330bd2fa96d74233036f6637e68262bd9f93))
* **server/middlewares:** :sparkles: htmx middlewares ([2731a9f](https://github.com/joshuar/go-hass-agent/commit/2731a9f5e90a8431bcdbef621b0c51d6145aa05c))
* **web/content:** :bento: add more favicon variants ([c2a61fe](https://github.com/joshuar/go-hass-agent/commit/c2a61fe0a5db2c85ee787306ba7b47b6e4e44527))
* **web/content:** :bento: bundle inter fonts ([41e39ab](https://github.com/joshuar/go-hass-agent/commit/41e39ab7a4aaf63dc37e9560282372c8de4c306b))
* **web/templates:** :sparkles: standard page template ([d26ffaf](https://github.com/joshuar/go-hass-agent/commit/d26ffaf72e16f092858e54f5aa427f69f64ce302))


### Bug Fixes

* **agent:** :bug: don't nest script sensor attributes, allow native format ([d202827](https://github.com/joshuar/go-hass-agent/commit/d2028271f31a5b47d1b554e888cd79f668c83bd5))
* **agent:** :bug: don't try to close channel that gets closed ([3ebd3f0](https://github.com/joshuar/go-hass-agent/commit/3ebd3f019d1f097df440cb321cba0b4179d173c5))
* **agent:** :bug: make sure context is set up correctly ([41efeb8](https://github.com/joshuar/go-hass-agent/commit/41efeb8aa6c39346862ca3569fe61dc38cedd5bd))
* **agent:** :fire: remove debugging output ([b3d67cc](https://github.com/joshuar/go-hass-agent/commit/b3d67cc499d8c598bca128df314b8fa7e556c41b))
* **assets:** :bug: remove deprecated `--terminal` command-line argument in systemd service file ([74f819d](https://github.com/joshuar/go-hass-agent/commit/74f819d97bb004bbc4325c6a0d21a1c68c6a0384))
* **config:** :bug: fix setting a custom path to the config ([3e21e84](https://github.com/joshuar/go-hass-agent/commit/3e21e84d718df1957b734abffc499848299bac73))
* **linux/desktop:** :bug: fix nil pointer references for desktop preferences ([b5f0130](https://github.com/joshuar/go-hass-agent/commit/b5f013069bae9dc2167931235ab3413199ef0e68))
* **linux/media:** :bug: fix getting media status ([c441792](https://github.com/joshuar/go-hass-agent/commit/c4417927ec94f88701227f002cf4114de15db297))
* **linux/net:** :bug: use device filters from preferences for netlink worker (link/address state sensors) ([521a101](https://github.com/joshuar/go-hass-agent/commit/521a1018c44f97cb449f910621037c985cfa72be))
* **web/assets:** :bug: fix path to templates ([abbd452](https://github.com/joshuar/go-hass-agent/commit/abbd452201fb6cf5eefdf75fdbb51f66f6e73ecc))
* **web/templates:** :bug: registration fixes ([c4dd8e3](https://github.com/joshuar/go-hass-agent/commit/c4dd8e3efa8315359ca5bdd7195e9f2fd78bbb54))


### Miscellaneous Chores

* release 14.0.0 ([1b73fdf](https://github.com/joshuar/go-hass-agent/commit/1b73fdf76cfdc7ca35a02d4511b256337a6aca67))

## [13.3.3](https://github.com/joshuar/go-hass-agent/compare/v13.3.2...v13.3.3) (2025-08-16)


### Bug Fixes

* **linux/disk:** :wrench: don't generate disk sensors for podman overlay mounts by default ([3e5e8d9](https://github.com/joshuar/go-hass-agent/commit/3e5e8d9a9a25b7d605d7a8d81a9f7e292a9665db))
* **linux/net:** :bug: use ignored devices for network rate sensors ([18e7a67](https://github.com/joshuar/go-hass-agent/commit/18e7a67419c714099e472db666db93914abc1569))
* **linux/net:** :wrench: ignore virtual network devices created by common container engines by default ([65d7688](https://github.com/joshuar/go-hass-agent/commit/65d768865f6c96bb8ea1ea7ed05c0e4adb87c399))

## [13.3.2](https://github.com/joshuar/go-hass-agent/compare/v13.3.1...v13.3.2) (2025-08-09)


### Bug Fixes

* **linux/disk:** :bug: use correct value for SATA disk attributes ([22920ac](https://github.com/joshuar/go-hass-agent/commit/22920ac0545c9804103fc350377f359aabd1f886))

## [13.3.1](https://github.com/joshuar/go-hass-agent/compare/v13.3.0...v13.3.1) (2025-07-27)


### Bug Fixes

* **hass:** :bug: don't try to update the config initially if agent is unregistered ([a468499](https://github.com/joshuar/go-hass-agent/commit/a4684999ff80515d5100afc758a77bd432828235))
* **hass:** :bug: fail fast on update job if not registered ([5c08685](https://github.com/joshuar/go-hass-agent/commit/5c086851b0e9a861637d238f13651b3682672f3b))

## [13.3.0](https://github.com/joshuar/go-hass-agent/compare/v13.2.8...v13.3.0) (2025-07-20)


### Features

* **linux/disk:** :sparkles: basic SMART disk monitoring sensors ([17e2b8f](https://github.com/joshuar/go-hass-agent/commit/17e2b8fca820cca7133a58dfdc71a4cb69073d5b))


### Reverts

* **linux:** :rewind: simplify capabilities checks ([c420e63](https://github.com/joshuar/go-hass-agent/commit/c420e63955a75e4567ec27776d48651f5b98b8b0))

## [13.2.8](https://github.com/joshuar/go-hass-agent/compare/v13.2.7...v13.2.8) (2025-07-05)


### Bug Fixes

* **hass:** :bug: correct detection of disabled entities and quieter logging of entity state changes ([dbd8d65](https://github.com/joshuar/go-hass-agent/commit/dbd8d65e278e2e9b406fecff6894c76d89ead76f))
* **linux/system:** :bug: ensure details are added when logging a D-Bus command execution ([763801c](https://github.com/joshuar/go-hass-agent/commit/763801c8a49c79dc8d6503182ae32d194142915b))

## [13.2.7](https://github.com/joshuar/go-hass-agent/compare/v13.2.6...v13.2.7) (2025-05-24)


### Bug Fixes

* **mqtt:** :bug: ensure a default server and topic are set if not passed in when configuring MQTT from the command-line ([beee456](https://github.com/joshuar/go-hass-agent/commit/beee456f8bc596c1f72b4efe80ba1fa7d903c894))
* **mqtt:** :bug: make sure devices use a unique "app" id for mqtt topics ([5b64ae9](https://github.com/joshuar/go-hass-agent/commit/5b64ae9ccc07e6771af263cd3e72924b61f45322))

## [13.2.6](https://github.com/joshuar/go-hass-agent/compare/v13.2.5...v13.2.6) (2025-05-18)


### Performance Improvements

* **linux:** :zap: use a shared pipewire monitor instance for workers monitoring pipewire events ([972e595](https://github.com/joshuar/go-hass-agent/commit/972e5957b5235723ecfe06749c6f76b0355ebf4c))

## [13.2.5](https://github.com/joshuar/go-hass-agent/compare/v13.2.4...v13.2.5) (2025-04-29)


### Bug Fixes

* **hass/api:** :ambulance: re-enable and fix websocket connection for notifications ([1914257](https://github.com/joshuar/go-hass-agent/commit/191425772245d6636c4679acf7d56284a663b300))
* **linux:** :bug: correct calculation for rate sensors ([07a3aa9](https://github.com/joshuar/go-hass-agent/commit/07a3aa99b5c40f77b95942106a5d79a88c480627))

## [13.2.4](https://github.com/joshuar/go-hass-agent/compare/v13.2.3...v13.2.4) (2025-04-12)


### Bug Fixes

* :bug: fix permissions ([04d201a](https://github.com/joshuar/go-hass-agent/commit/04d201a1ee644070d0496c60a27e37215e9508e1))
* **commands:** :bug: commands worker doesn't need to implement a worker with preferences ([f537cf5](https://github.com/joshuar/go-hass-agent/commit/f537cf5a1555372f183df6d78953a2a89a4aa82c))
* **dbusx:** :bug: improve watching multiple methods ([b6e8e4b](https://github.com/joshuar/go-hass-agent/commit/b6e8e4b863779f6e5102c123d47af4bbd8135b13))
* **linux/net:** :bug: correct state for networkmanager connection states ([f7f7a3b](https://github.com/joshuar/go-hass-agent/commit/f7f7a3b44eedb24c50452362daa52efb1e6684c5))
* **mqtt:** :sparkles: all MQTT functionality can now be disabled in preferences ([0e72e47](https://github.com/joshuar/go-hass-agent/commit/0e72e47aff011e68e785fef9bbac2f9a6c51fa8d))
* **scripts:** :sparkles: scripts worker can now be disabled in preferences ([7912047](https://github.com/joshuar/go-hass-agent/commit/7912047280257d6aab69a64d359d6fc2c4679a9f))
* **workers:** :bug: improved worker context handling ([6389dbc](https://github.com/joshuar/go-hass-agent/commit/6389dbcfa8371076eb655b3a6268c0f634ce56e6))


### Performance Improvements

* :zap: remove unnecessary abstractions ([259322e](https://github.com/joshuar/go-hass-agent/commit/259322ef7b1779e4097210f623e669a36f1cd225))
* **workers:** :zap: better delta calculation ([6dd9cd2](https://github.com/joshuar/go-hass-agent/commit/6dd9cd28c3ba4dd3eb5f4accfa421ced9a4f5d2d))

## [13.2.3](https://github.com/joshuar/go-hass-agent/compare/v13.2.2...v13.2.3) (2025-03-22)


### Bug Fixes

* **linux:** :bug: don't try to run nil workers ([6c6fcec](https://github.com/joshuar/go-hass-agent/commit/6c6fcec9d6bd8fdc7b8bc0e19e6d47c921421276))
* **linux:** :bug: don't try to run nil workers ([db19d08](https://github.com/joshuar/go-hass-agent/commit/db19d08dcb43cead345180ba4d8c22f19223939d))

## [13.2.2](https://github.com/joshuar/go-hass-agent/compare/v13.2.1...v13.2.2) (2025-03-22)


### Performance Improvements

* **linux/media:** :zap: improved pipewire monitoring for webcam/mic in use sensors ([2b65d49](https://github.com/joshuar/go-hass-agent/commit/2b65d49c7b448f8defaba2a42eb0caef4f2b6b6a))

## [13.2.1](https://github.com/joshuar/go-hass-agent/compare/v13.2.0...v13.2.1) (2025-03-15)


### Bug Fixes

* **device:** :bug: don't try to add mqtt workers that didn't fail and cannot be enabled ([e7c1935](https://github.com/joshuar/go-hass-agent/commit/e7c1935a29d42385cc1c9af605d92dabc526061e))
* **linux/power:** :bug: graceful shutdown of screen lock sensor worker ([b089948](https://github.com/joshuar/go-hass-agent/commit/b089948960beed1db83a561944805c136f062925))

## [13.2.0](https://github.com/joshuar/go-hass-agent/compare/v13.1.1...v13.2.0) (2025-03-15)


### Features

* **app:** :recycle: migrate to modular app design ([c15b6db](https://github.com/joshuar/go-hass-agent/commit/c15b6dbb8c70ec1d8bc3424ad66a4e882b568a21))


### Bug Fixes

* **device:** :bug: use config rest API endpoint for connection latency sensor to eliminate log spam ([6be1fd5](https://github.com/joshuar/go-hass-agent/commit/6be1fd5fe7c0dcedc67ecb0b62ae917bb15278f9))
* **linux:** :bug: mqtt workers should create their own channels for messages ([f3e1b0e](https://github.com/joshuar/go-hass-agent/commit/f3e1b0e2c9242c70ae4396ec6eed11dfcbd0a439))
* **mqtt:** :bug: correctly parse numeric values from toml ([503d645](https://github.com/joshuar/go-hass-agent/commit/503d645cbc1f91920836c63385d028f219cc3723))
* **mqtt:** :fire: remove spew ([1ec500e](https://github.com/joshuar/go-hass-agent/commit/1ec500eb6ae38e5cd4acd619f263a75e8208ee12))
* **workers:** :zap: improve worker shutdown handling ([2290564](https://github.com/joshuar/go-hass-agent/commit/2290564eca19b6050a475b313bea15367b2153c4))

## [13.1.1](https://github.com/joshuar/go-hass-agent/compare/v13.1.0...v13.1.1) (2025-03-02)


### Bug Fixes

* **hass:** :bug: ensure entity data is writeable before marshaling from event ([9c105c2](https://github.com/joshuar/go-hass-agent/commit/9c105c2dff20ec06747a2bc1f77b43eb756cb1ac))
* **linux/power:** :bug: ignore non power state signals in power state sensor code ([2f9335c](https://github.com/joshuar/go-hass-agent/commit/2f9335c38d2331135c151dad43319fa30c5c2f66))
* **linux/system:** :bug: correct logic around session added/removed events ([ded1d58](https://github.com/joshuar/go-hass-agent/commit/ded1d58548e10edf6e21759061e0c9f0f4340106))

## [13.1.0](https://github.com/joshuar/go-hass-agent/compare/v13.0.1...v13.1.0) (2025-03-02)


### Features

* **hass:** :building_construction: perform validation on data received before making requests ([96cc423](https://github.com/joshuar/go-hass-agent/commit/96cc423c81a09fe1acd1d8243940ff349f57db0b))
* **service:** :sparkles: systemd service improvements ([bcae15f](https://github.com/joshuar/go-hass-agent/commit/bcae15f2cd4fa76eac092d825c083bf39154cb82))


### Bug Fixes

* **hass:** :bug: retryable is a required field on entities and requests ([9677b9c](https://github.com/joshuar/go-hass-agent/commit/9677b9ca1c7b2532e09f8685fd060a5189689fc7))

## [13.0.1](https://github.com/joshuar/go-hass-agent/compare/v13.0.0...v13.0.1) (2025-03-02)


### Bug Fixes

* **linux/media:** :bug: restart pipewire monitor if it crashes unexpectedly ([351e8b2](https://github.com/joshuar/go-hass-agent/commit/351e8b29a379360aa9f8f8b72413d72bb0c5632b))
* **linux/power:** :bug: don't send nil power state entity ([f8c4837](https://github.com/joshuar/go-hass-agent/commit/f8c483757683e404a6ba5dc781bf4ed080a1fc8a))

## [13.0.0](https://github.com/joshuar/go-hass-agent/compare/v12.0.1...v13.0.0) (2025-03-01)


### ⚠ BREAKING CHANGES

* **scripts:** script sensors are now run using a new scheduler backend. The scheduling is still based on cron schedule strings but some more complex/esoteric schedule strings may no longer work.

### Features

* :sparkles: add a job scheduler ([281c03c](https://github.com/joshuar/go-hass-agent/commit/281c03cd23d56e2382f5b0cf8070d17a4e5b3009))
* **linux:** :zap: use a common struct to manage rate sensor values ([7bbbedb](https://github.com/joshuar/go-hass-agent/commit/7bbbedb2ba5a3c33bbd566b05d2c68c9a907c436))
* **linux/media:** :sparkles: add webcam and mic in use sensors ([46f2ad8](https://github.com/joshuar/go-hass-agent/commit/46f2ad8825ea399eb60893800f8069612f90d7e6))
* **models:** :sparkles: start implementing various types and objects with openapi-codegen ([ba96a3e](https://github.com/joshuar/go-hass-agent/commit/ba96a3e221d0a2d33ced5d095eedb493fbc7712a))
* **scheduler:** :sparkles: implement a poll trigger with jitter ([26b3d99](https://github.com/joshuar/go-hass-agent/commit/26b3d99ebd2b9f002423fdf4cc7ae3160dc2cd21))
* **scripts:** :boom: use scheduler for scripts ([1c4e8dd](https://github.com/joshuar/go-hass-agent/commit/1c4e8dd00d49e9985c1f841fad275c553d40c5c5))


### Bug Fixes

* :fire: fix missed worker conversions ([6b37417](https://github.com/joshuar/go-hass-agent/commit/6b3741749557f4b1b43226b33654b5e57aeb9f7e))
* :rotating_light: fix more linter warnings ([a563eec](https://github.com/joshuar/go-hass-agent/commit/a563eecf3f0cf24e22809a541b49fb3cd79bca54))
* **hass:** :bug: actually update local registration on sensor registration ([17e8966](https://github.com/joshuar/go-hass-agent/commit/17e8966bf6be21cb1ce4e468b0c106af7a34829a))
* **hass:** :bug: better handling of initial case where HA config has no entity status ([7156ec9](https://github.com/joshuar/go-hass-agent/commit/7156ec90476f4968d8648bb849a3655d9cc0cb95))
* **hass:** :bug: ensure authorization request header is set and url format is valid ([6969547](https://github.com/joshuar/go-hass-agent/commit/6969547aaf9c0832c944148fb8a6a375b7523ced))
* **hass/api:** :bug: correct display of error code returned from Home Assistant Rest API ([82966ab](https://github.com/joshuar/go-hass-agent/commit/82966ab62dfc7d036d662f92d8d47836591bd801))
* **linux/battery:** :rotating_light: clean-up code from linter warnings ([b901c65](https://github.com/joshuar/go-hass-agent/commit/b901c658e8c1c03cd6dec8e36f3cf28cebd4a5b8))
* **linux/cpu:** :rotating_light: clean-up code from linter warnings ([8dda011](https://github.com/joshuar/go-hass-agent/commit/8dda0115aeab658c608368554845fe9c6e2a00bd))
* **linux/desktop:** clean-up code from linter warnings ([be848db](https://github.com/joshuar/go-hass-agent/commit/be848dbdfe4018685d1dd7acd81bb58c393f1ca3))
* **linux/media:** :bug: fix preferences location for media sensors ([d8696b1](https://github.com/joshuar/go-hass-agent/commit/d8696b159bec4e566c197c1b93c69f4c8b61be39))
* **models:** :bug: make sure device/state class are sent as strings to Home Assistant ([5c83f9c](https://github.com/joshuar/go-hass-agent/commit/5c83f9cd1d4d66e96c13564dbcb0610cf716943e))


### Performance Improvements

* **agent:** :recycle: combine sensor and event workers ([b35bf98](https://github.com/joshuar/go-hass-agent/commit/b35bf98fb3e2637625633b2997ab2a4bed512d9e))
* **hass:** :zap: reduce number of requests per entity update ([1ab7fad](https://github.com/joshuar/go-hass-agent/commit/1ab7fadbfdf3bb40476e0b0a30286daf1cf087f0))

## [12.0.1](https://github.com/joshuar/go-hass-agent/compare/v12.0.0...v12.0.1) (2025-02-05)


### Bug Fixes

* :bug: improve error handling in workers ([af3c485](https://github.com/joshuar/go-hass-agent/commit/af3c485cb96ea581df6a7e107410fafe15ccc297))
* **linux/disk:** :bug: don't send disk usage sensors with invalid values ([05bc753](https://github.com/joshuar/go-hass-agent/commit/05bc75348236f0b973c9ee83aac1fc49a9ca1052))

## [12.0.0](https://github.com/joshuar/go-hass-agent/compare/v11.1.2...v12.0.0) (2025-02-02)


### ⚠ BREAKING CHANGES

* **components/preferences:** Sensors and (MQTT) controls preferences have been restructured in the preferences file. Users who have customised any sensor/control preferences will need to manually migrate the changes to the new structure.

### Features

* :sparkles: add `--path` command-line option for specifying a path for preferences/logs/data ([85b13b3](https://github.com/joshuar/go-hass-agent/commit/85b13b3b5aa2f912acfbfb6f4121f10f723f7e6b))
* **agent/sensor:** :sparkles: add preference to disable connection latency sensor ([dc31bed](https://github.com/joshuar/go-hass-agent/commit/dc31bed70767e36a072f184d8b963f15a0dbbc9b))
* **agent/sensor:** :sparkles: add preference to disable version sensor ([dabe732](https://github.com/joshuar/go-hass-agent/commit/dabe732b5f4ff6a0db55b9637e44f5ee8cbfa167))
* **components/preferences:** restructure sensor/control preferences ([5376ce1](https://github.com/joshuar/go-hass-agent/commit/5376ce1eed04763a5bf3f54ae017779f403e7077))
* **linux/battery:** :sparkles: add preference to disable battery sensors ([039791e](https://github.com/joshuar/go-hass-agent/commit/039791eca0d1a928c5fb2b4d2c9648004b35e8cf))
* **linux/battery:** :sparkles: break out voltage and energy battery attributes into their own sensors ([5875192](https://github.com/joshuar/go-hass-agent/commit/58751920e7a0febe929d2fd6540b1b999d1a46e1))
* **linux/cpu:** :sparkles: add preferences to disable cpu load avgs and vulnerabilities sensors ([aa46a99](https://github.com/joshuar/go-hass-agent/commit/aa46a9971f9f38d5aa7249c2c87b6deeb1efad1e))
* **linux/desktop:** :sparkles: add preference to disable desktop sensors ([cba5a50](https://github.com/joshuar/go-hass-agent/commit/cba5a5047475b61a674a68c0388ace732d87bbde))
* **linux/disk:** :sparkles: add preferences to disable disk io/usage sensors and set intervals for sensor updates ([96be29b](https://github.com/joshuar/go-hass-agent/commit/96be29b07c19fbe334146a7a30fb2ac5eca907b1))
* **linux/location:** :sparkles: add preference to disable location tracking ([ef6a113](https://github.com/joshuar/go-hass-agent/commit/ef6a113650f47cd87ed7df962c0bfcfdc003f257))
* **linux/media:** :sparkles: add preferences to disable mpris, audio sensors and controls ([8eedf6d](https://github.com/joshuar/go-hass-agent/commit/8eedf6daccbf132201bcab185c1186feb66351c5))
* **linux/mem:** :sparkles: add preferences for disabling and setting update interval of mem usage sensors and disabling oom events tracking ([83e1efc](https://github.com/joshuar/go-hass-agent/commit/83e1efc0e3aabebecf64ae508c8ac7f5f0c83009))
* **linux/net:** :sparkles: add preference to set the interval for network device rates sensor updates ([97c104f](https://github.com/joshuar/go-hass-agent/commit/97c104fb49794e2a85e120581915eb601f7058e8))
* **linux/power:** :sparkles: add an MQTT powered control to set a sleep/shutdown inhibit lock ([7c63100](https://github.com/joshuar/go-hass-agent/commit/7c63100e888611f45eeb7f52a533302309224f3a))
* **linux/power:** :sparkles: add preferences for disabling the various power-based sensors individuallly ([6ca6fe1](https://github.com/joshuar/go-hass-agent/commit/6ca6fe1795b97db687c34f1c3ae103578e2debb7))
* **linux/system:** :sparkles: add preferences for disabling/setting poll intervals (where appropriate) for all system sensor and event workers ([cca049b](https://github.com/joshuar/go-hass-agent/commit/cca049be2d01d49ae5ecb3e12274222ee61b5e13))


### Bug Fixes

* :fire: remove spew ([27fed4a](https://github.com/joshuar/go-hass-agent/commit/27fed4a1d85532de94a51be54f78f98b837f40fd))
* **components/preferences:** :bug: additional logic fixes ([a7b138e](https://github.com/joshuar/go-hass-agent/commit/a7b138e584d5f11f89631ff45481bd5c19f30aed))
* **components/preferences:** :bug: parse existing worker preferences correctly ([39186d2](https://github.com/joshuar/go-hass-agent/commit/39186d2c1f0a4d2a7df215b7146b50cb6f5de50d))
* **dbusx:** :bug: better user session finder ([dbb5e06](https://github.com/joshuar/go-hass-agent/commit/dbb5e0684bcf8cdb17550da0c9fca4ef4600b081))
* **hass:** :bug: ensure registration server and token are set in preferences ([7b9c2fa](https://github.com/joshuar/go-hass-agent/commit/7b9c2fa68db4990efbd1d5010fe5251fcd2accda))
* **hass/api:** :bug: remove regression where websocket url retained any port element (spoiler: it shouldn't) ([4d3a4bc](https://github.com/joshuar/go-hass-agent/commit/4d3a4bc580e14008cd7b916cbb38eca08c078e16))
* **hass/discovery:** :bug: remove regression where the default server was not listed on discovery of servers during graphical registration ([c5639d0](https://github.com/joshuar/go-hass-agent/commit/c5639d0ce5bd7d3c3cd69ccdfa3609455111021f))
* **linux:** :bug: ensure polling sensors use poll interval from preferences, default otherwise ([79092be](https://github.com/joshuar/go-hass-agent/commit/79092be5cb2e4ba4392da9af0f8189e1d87cb9c7))
* **linux/cpu:** :bug: actually add units for 114a35fa93c158ffe5956a94d6885e840e6a2465 ([068828b](https://github.com/joshuar/go-hass-agent/commit/068828bf0879899040b2b9fec6a612f1f235526d))
* **linux/cpu:** :bug: add units to cpu usage count sensors ([114a35f](https://github.com/joshuar/go-hass-agent/commit/114a35fa93c158ffe5956a94d6885e840e6a2465))
* **linux/desktop,linux/battery:** :fire: remove debugging output ([530e8c2](https://github.com/joshuar/go-hass-agent/commit/530e8c233dff980cbf47033020c0bae289397e7f))
* **linux/location:** :bug: correct type conversion ([e0ade99](https://github.com/joshuar/go-hass-agent/commit/e0ade997be11b085ce5900685f03078741a9ca63))
* **linux/mem:** :bug: fix linter warnings in 83e1efc0e3aabebecf64ae508c8ac7f5f0c83009 ([3acd7f6](https://github.com/joshuar/go-hass-agent/commit/3acd7f6b486bbfc48ad1a00d67effcaf93512791))
* **linux/power:** :rotating_light: fix linter warning ([e38c259](https://github.com/joshuar/go-hass-agent/commit/e38c259d265687bba1c10cb38d041a71a29a583c))
* **pkg/whichdistro:** :bug: ignore lines that are not key=value pairs ([ef90391](https://github.com/joshuar/go-hass-agent/commit/ef90391faeceeaef7f52acf6cb0d6fb5b490851a))
* **preferences:** :bug: ensure version is written to `preferences.toml` when it is saved ([7b720b4](https://github.com/joshuar/go-hass-agent/commit/7b720b45c535e9309ea3ccd854a0a338065041fc))
* **scripts:** :bug: actually warn about script parsing errors ([101b2a8](https://github.com/joshuar/go-hass-agent/commit/101b2a8a7534e35976237c0dae83a5517f52cac7))
* **scripts:** :bug: remove regression whereby script sensors were not sending their sensor states initially at agent start-up ([cbb0344](https://github.com/joshuar/go-hass-agent/commit/cbb03441f2c4d690088eee013aa8395afc70e7da))
* **ui:** :bug: store/fetch mqtt preferences from context ([a20f509](https://github.com/joshuar/go-hass-agent/commit/a20f509aac496eb31888d6b1c613e425af267d92))


### Performance Improvements

* **agent:** :zap: load up the worker context once and pass to all worker processes ([7dd5c57](https://github.com/joshuar/go-hass-agent/commit/7dd5c57c0ffdd3a91ffe422f0dbf65b5305008b0))

## [11.1.2](https://github.com/joshuar/go-hass-agent/compare/v11.1.1...v11.1.2) (2025-01-11)


### Bug Fixes

* **agent:** :bug: new graphical registration flow ([5674352](https://github.com/joshuar/go-hass-agent/commit/5674352c0cce9f390675a5efc9a1163a562ecc4a))

## [11.1.1](https://github.com/joshuar/go-hass-agent/compare/v11.1.0...v11.1.1) (2025-01-11)


### Bug Fixes

* **agent:** :bug: detect more types of "laptop" chassis ([f6f32fd](https://github.com/joshuar/go-hass-agent/commit/f6f32fda3de5d462523c11e0deac12b89c6e219b))
* **agent/sensor:** :bug: don't add nonexistent agent workers to worker list ([3d35099](https://github.com/joshuar/go-hass-agent/commit/3d35099e39eced7f0cd6c8fc343f49d074883caf))
* **linux/mem:** :bug: correct unit for percentage memory usage sensors ([aa59eee](https://github.com/joshuar/go-hass-agent/commit/aa59eee151dbb80c0bf3ff25d2bd1fa44d0adb78))

## [11.1.0](https://github.com/joshuar/go-hass-agent/compare/v11.0.0...v11.1.0) (2025-01-06)


### Features

* **linux/power:** :sparkles: user enum sensor device class for power state sensor ([f7c0bd6](https://github.com/joshuar/go-hass-agent/commit/f7c0bd627fb9526cc648414f222f924b0ebf3a51))


### Bug Fixes

* **preferences:** :bug: actually use default preferences if no preferences file is found ([70f49ef](https://github.com/joshuar/go-hass-agent/commit/70f49ef06f1c6686ffabdf49cb1dcb6476b1402a))
* **preferences:** :bug: more preferences fixes after 7591c7a ([ee5f2d6](https://github.com/joshuar/go-hass-agent/commit/ee5f2d6b3a25cce989900c0ff917127b9a5b77d6))

## [11.0.0](https://github.com/joshuar/go-hass-agent/compare/v10.5.1...v11.0.0) (2024-12-31)


### ⚠ BREAKING CHANGES

* **preferences:** Worker preferences are now in the agent preferences file, under a "worker_prefs" section. Any existing custom preferences will need to be manually migrated to this file.

### Features

* :sparkles: allow disabling app sensors ([71a4969](https://github.com/joshuar/go-hass-agent/commit/71a49696dc7fff379a78e039099411842428f269))
* **hass,linux:** :recycle: support flagging for retryable requests through sensor options ([5d55cc6](https://github.com/joshuar/go-hass-agent/commit/5d55cc60e43831f3407d065901588d732e6986c4))
* **hass,linux:** :sparkles: use options pattern to create sensors ([b614ec3](https://github.com/joshuar/go-hass-agent/commit/b614ec3f7b02e3595c19a1dacefd255add6873c9))
* **hass:** :sparkles: add support to allow some requests to be retried ([b103679](https://github.com/joshuar/go-hass-agent/commit/b10367941aa487f6ee7cf9eaaa8179d51121a133))
* **hass:** :sparkles: use options pattern to create sensor requests ([73f218b](https://github.com/joshuar/go-hass-agent/commit/73f218b1a7c2b4e2bc2e696f1e5934d497182bec))
* **linux/cpu,linux/system:** :sparkles: add ability to specify update interval for cpu and hwmon sensor polling ([7f8450e](https://github.com/joshuar/go-hass-agent/commit/7f8450efd1a203eee4c03eaa41e7259888652d54))
* **linux/cpu:** :truck: split cpu usage and frequency workers ([cc18b67](https://github.com/joshuar/go-hass-agent/commit/cc18b676be949f1bd6102d2f9305c680828e4424))
* **linux/power:** :sparkles: power state and screen lock sensor requests will be retried on response failure ([e4ca6e7](https://github.com/joshuar/go-hass-agent/commit/e4ca6e74a6009e82b5315169537f35c57a192b77))
* **preferences:** :sparkles: validate worker preferences when loading and use defaults if invalid ([748b48f](https://github.com/joshuar/go-hass-agent/commit/748b48fec5790088bba98444f4b56d1c997d50ef))


### Bug Fixes

* :bug: code cleanup missed in 7591c7ac1123bf409144c650a5cd47f8eb49ee07 ([f6dca52](https://github.com/joshuar/go-hass-agent/commit/f6dca52870a20ef1459e5a57e3cac9fd46d7919b))
* **agent:** :bug: fix registration flow from changes in 7591c7a ([70176ce](https://github.com/joshuar/go-hass-agent/commit/70176ceababf33926023a8d0d073e15d1a8d6daf))
* **hass:** :rotating_light: fix linter issues ([92b82d0](https://github.com/joshuar/go-hass-agent/commit/92b82d0f026e943333c0cb02718d25039b2013d1))
* **linux/battery:** :bug: don't add already tracked batteries ([76b78e4](https://github.com/joshuar/go-hass-agent/commit/76b78e4bc9ff16f6209b789ecebff52ba22b7ca1))
* **preferences:** :bug: ensure consistent naming of preferences through using string constants ([1397c4a](https://github.com/joshuar/go-hass-agent/commit/1397c4a9e188f37187f2f672b70788f6f420e067))


### Performance Improvements

* **agent:** :zap: don't use a "fat context" for agent options ([1e9d3c9](https://github.com/joshuar/go-hass-agent/commit/1e9d3c988b3de2fcce9bfe71c26793619a797b0f))


### Code Refactoring

* **preferences:** :recycle: merge worker and agent preferences into single file ([7591c7a](https://github.com/joshuar/go-hass-agent/commit/7591c7ac1123bf409144c650a5cd47f8eb49ee07))

## [10.5.1](https://github.com/joshuar/go-hass-agent/compare/v10.5.0...v10.5.1) (2024-11-20)


### Bug Fixes

* **linux/mem:** :fire: remove debug output ([ec1b975](https://github.com/joshuar/go-hass-agent/commit/ec1b9750f05ce05bec7989b44484f057fa49d391))

## [10.5.0](https://github.com/joshuar/go-hass-agent/compare/v10.4.0...v10.5.0) (2024-11-19)


### Features

* **linux/cpu:** :sparkles: add preferences to optionally disable all cpu (and specifically, cpu frequency) sensors ([ecc5cc6](https://github.com/joshuar/go-hass-agent/commit/ecc5cc6a6a6e3d7615ef24fe7adac0811983c5d5))
* **linux/mem:** :sparkles: send oom events to Home Assistant ([c491e81](https://github.com/joshuar/go-hass-agent/commit/c491e810f21c574fb784af9babc607416d6ba8aa))
* **linux/system:** :sparkles: add preferences to optionally disable hwmon sensors ([7b65aab](https://github.com/joshuar/go-hass-agent/commit/7b65aab7b6473c2377c19af7e1206241f2765954))


### Bug Fixes

* **hass:** :bug: correct JSON marshaling ([bd31214](https://github.com/joshuar/go-hass-agent/commit/bd31214b7304e60483d8166070901c11a4a30b0c))
* **linux/cpu:** :bug: cpu process state counts should not be totalincreasing state class ([0183281](https://github.com/joshuar/go-hass-agent/commit/0183281cd03514ab4a274548a4058f6fcf0a6613))
* **linux/net:** :bug: only use link up/down/invalid netlink messages for link state ([27919b4](https://github.com/joshuar/go-hass-agent/commit/27919b46fbba0af2ba043a6ea694e84bbe00cdbb))
* **linux/net:** :bug: treat unknown link state as down state ([813485e](https://github.com/joshuar/go-hass-agent/commit/813485e73b56de1605d5ed23ee80583d0599bad2))

## [10.4.0](https://github.com/joshuar/go-hass-agent/compare/v10.3.2...v10.4.0) (2024-10-31)


### Features

* **agent:** :sparkles: add an interface to represent a worker with preferences for future use ([446857e](https://github.com/joshuar/go-hass-agent/commit/446857e0ccdae81e5834e9223c21aaa791fd3b90))
* **agent:** :sparkles: implement event controller for event workers in agent ([c1d2033](https://github.com/joshuar/go-hass-agent/commit/c1d20330d8a9720f9299d34d46663cd09398b8cf))
* **agent/sensor:** :sparkles: add preference to disable external ip sensor if desired ([17f8d97](https://github.com/joshuar/go-hass-agent/commit/17f8d97ddf7a3354df9d637390c2f949ceda66e5))
* **hass:** :sparkles: add support for sending events to Home Assistant ([6debf7e](https://github.com/joshuar/go-hass-agent/commit/6debf7ec835a94cda4cc91b91066094a91d62320))
* **linux:** :sparkles: add session events ([61b87e6](https://github.com/joshuar/go-hass-agent/commit/61b87e6b876e2c72c0b13b683e6b953890cc9b5e))
* **linux:** :sparkles: add tracking stats from chronyd as sensors ([3de2c09](https://github.com/joshuar/go-hass-agent/commit/3de2c0946b1149aede5227da6d11678894048306))
* **linux:** :sparkles: add user preference to define devices to ignore for network rates sensors ([c36e14e](https://github.com/joshuar/go-hass-agent/commit/c36e14eed3e0d0126195a96ff91bd50e860eb020))
* **linux/media:** :sparkles: support user preferences for camera worker ([fef5dd9](https://github.com/joshuar/go-hass-agent/commit/fef5dd979e958903d67e94b7a4cf930b2244e846))
* **linux/net:** :sparkles: filter on user-defined network devices for networkmanger connection state sensors ([77e3372](https://github.com/joshuar/go-hass-agent/commit/77e337233ac47213a93acdf8533e3d01c38ebe03))
* **preferences:** :sparkles: provide a worker preference to completely disable the worker (and its sensors/events/controls) ([23c940b](https://github.com/joshuar/go-hass-agent/commit/23c940bf401721d6d75e22e16f2713c1ca36e8ee))
* **preferences:** :sparkles: support worker preferences ([c7c49ac](https://github.com/joshuar/go-hass-agent/commit/c7c49ac86c0bd44eaf06f1b442497bbe94fd8203))


### Bug Fixes

* **hass:** :bug: fix validation of event requests ([d95955e](https://github.com/joshuar/go-hass-agent/commit/d95955eaebae0e5c44f6185607c7f68eee2508bb))
* **hass:** :bug: rework marshaling of sensor requests ([1c579d9](https://github.com/joshuar/go-hass-agent/commit/1c579d964dfce6f51643ca7276baa8e4bcb72093))
* **linux:** :bug: for ignored devices, ensure their stats are still tracked as part of the total network rate sensors ([468f692](https://github.com/joshuar/go-hass-agent/commit/468f6924cfc6761946a26a6b686478beee953d89))


### Performance Improvements

* :zap: share a validator instance across packages ([81bf9e8](https://github.com/joshuar/go-hass-agent/commit/81bf9e86ece23c0082a8ce7075e31a8aad8ae553))
* **agent:** :zap: rework controller/worker concept ([8bca59f](https://github.com/joshuar/go-hass-agent/commit/8bca59f0dc06a05f6b943b0c6794b68a9943e56b))
* **hass:** :zap: rework sending requests ([a8cd590](https://github.com/joshuar/go-hass-agent/commit/a8cd5909fa01a124b09d4d7649e245755ba7f517))

## [10.3.2](https://github.com/joshuar/go-hass-agent/compare/v10.3.1...v10.3.2) (2024-10-15)


### Bug Fixes

* **linux:** :bug: fix total calculation for network rates sensors ([f51acaa](https://github.com/joshuar/go-hass-agent/commit/f51acaac8314df34b9d3356633e73dee91457871))

## [10.3.1](https://github.com/joshuar/go-hass-agent/compare/v10.3.0...v10.3.1) (2024-10-12)


### Bug Fixes

* **linux:** :bug: safely check for hsi properties ([3bb3849](https://github.com/joshuar/go-hass-agent/commit/3bb3849bc802264c480c45ca9ab2f20ce290151f))


### Performance Improvements

* **container:** :zap: don't install mage for container build ([dace627](https://github.com/joshuar/go-hass-agent/commit/dace627addf2287307426d3d746fb8bf2876b358))
* **hass:** :zap: move validation of requests ([8d0eca5](https://github.com/joshuar/go-hass-agent/commit/8d0eca568b8962d5db38c667f84023f23c1c6f39))

## [10.3.0](https://github.com/joshuar/go-hass-agent/compare/v10.2.1...v10.3.0) (2024-10-01)


### Features

* **agent:** :sparkles: add connection latency sensor ([d55b1ed](https://github.com/joshuar/go-hass-agent/commit/d55b1ed552646abca51529f35595fbfca09bf3a6))
* **dbusx:** :sparkles: Add a Data type for fetching data via a D-Bus method ([edf80e1](https://github.com/joshuar/go-hass-agent/commit/edf80e1f4dd69e3b9a9ac052a633cc9678093405))
* **linux:** :sparkles: add a sensor to track if the kernel has reported any CPU vulnerabilities ([8d5ebf2](https://github.com/joshuar/go-hass-agent/commit/8d5ebf26306a0af1092380f1b8ecc00ffbbed4ef))
* **linux:** :sparkles: add link sensors ([cece6ed](https://github.com/joshuar/go-hass-agent/commit/cece6ede214f193ab2c7c509a05a070c0a31dfe1))
* **linux:** :sparkles: add per device network counts/rates sensors as well as the total counts/rates ([895125f](https://github.com/joshuar/go-hass-agent/commit/895125fb530e475086ffa1ce9b8dfec0c2c67a5a))
* **linux:** :sparkles: add sensor for displaying firmware security details ([dae37b4](https://github.com/joshuar/go-hass-agent/commit/dae37b448a7a1d3f1041b0c01b00aec3b4b2f43e))
* **linux:** :sparkles: add sensors for IO ops in progress per disk (and total of all disks) ([ea33a54](https://github.com/joshuar/go-hass-agent/commit/ea33a544edaf22159797290cf6e1e4fd96fb9937))
* **linux:** :sparkles: switch total cpu context switches and processes created sensors from totals to rates ([ed015e7](https://github.com/joshuar/go-hass-agent/commit/ed015e7cfaebcb6c4be06f0bc0cd6f060fdd6d01))


### Bug Fixes

* :rotating_light: add more nil pointer protections ([f1f4293](https://github.com/joshuar/go-hass-agent/commit/f1f429391fedaa6ec3f5bfebcf13d3dc1fca704a))
* **agent:** :bug: fix error handling and change endpoint for connection latency sensor ([6dedbc1](https://github.com/joshuar/go-hass-agent/commit/6dedbc13384ceb667e0bb5fb350747a08ee33876))
* **agent:** :bug: pass preferences to notifications worker ([30178cd](https://github.com/joshuar/go-hass-agent/commit/30178cd0722e6e97ecab194205032c39e2ee2160))
* **agent:** :bug: try to protect against empty response in connection latency sensor ([b40ccc7](https://github.com/joshuar/go-hass-agent/commit/b40ccc75692506cce899fbe16a7d0dd03384ba51))
* **agent:** :bug: uncomment commented block for testing ([9f4b656](https://github.com/joshuar/go-hass-agent/commit/9f4b65619b6a6b641b2ca7e19393e07dd9ca8d1a))
* **hass:** :bug: don't exclude nil value sensors when retrieving sensor list ([886b7eb](https://github.com/joshuar/go-hass-agent/commit/886b7eb8431aaaae6173b1c054d9b3f29e60567c))
* **hass:** :bug: simplify validation of sensor requests ([6db1638](https://github.com/joshuar/go-hass-agent/commit/6db1638fb0458e5d2edd55854c33aafaa39f2a02))
* **linux:** :art: better netlink shutdown handling in link sensor worker ([a265fec](https://github.com/joshuar/go-hass-agent/commit/a265fec5ea2bf811674286408ebb22f454f74e7d))
* **linux:** :bug: actually track running app and total running apps in worker ([66e4a19](https://github.com/joshuar/go-hass-agent/commit/66e4a198993392d8098d1557a5fedeab67d98f3e))
* **linux:** :bug: add missing disk IO sensor attribute so that disk read/write rates are calculated correctly ([8d7e6af](https://github.com/joshuar/go-hass-agent/commit/8d7e6af36111af2385d94e9fd5fc08b2b4382a3e))
* **linux:** :bug: add missing disk IO sensor attribute so that disk read/write sensors are calculated correctly ([9b024ee](https://github.com/joshuar/go-hass-agent/commit/9b024ee8a7ed72b512a131ff8e59a186e6745b33))
* **linux:** :bug: avoid pointer ref/deref ([86a5b5c](https://github.com/joshuar/go-hass-agent/commit/86a5b5c65fea60f5cdc64333f5d1ea7706ba39f2))
* **linux:** :bug: correct screen lock state with new device class ([f6811bb](https://github.com/joshuar/go-hass-agent/commit/f6811bb1ffdadca13a2710d61cd1c560f23e6bfc))
* **linux:** :bug: don't add `last_reset` attribute for cpu usage sensors with `total_increasing` state class ([89b903f](https://github.com/joshuar/go-hass-agent/commit/89b903f5e9383796a1556fb51ba4ce5722c426be))
* **linux:** :bug: event based workers should expose a send-only channel on Events method ([40e1751](https://github.com/joshuar/go-hass-agent/commit/40e1751525f6c45cca428ecc92688c319640f7cd))
* **linux:** :bug: filter all of `/run` from usage stats ([a0d57bf](https://github.com/joshuar/go-hass-agent/commit/a0d57bf7f2cfb4c96fecde0c058eef206941852f))
* **linux:** :bug: filter more mount points from generating usage sensors ([b238687](https://github.com/joshuar/go-hass-agent/commit/b238687acc956bd2a072ddb6a5e75257b76c3129))
* **linux:** :bug: fix changed network rates sensor types stringer ([64e9df9](https://github.com/joshuar/go-hass-agent/commit/64e9df942819ea7f56423db9ed115f725f69b064))
* **linux:** :bug: get the current screen lock state and send as a sensor on start ([40cbb57](https://github.com/joshuar/go-hass-agent/commit/40cbb57c322da2f7d6b0d2f6c99363158c6cf284))
* **linux:** :bug: protect against potential nil pointer exception ([ab99be0](https://github.com/joshuar/go-hass-agent/commit/ab99be0129089978dde4d8814bd57fd60360e169))
* **linux:** :bug: use distinct device classes for intrusion and alarm hardware sensors ([53b552b](https://github.com/joshuar/go-hass-agent/commit/53b552b2855b9eacf29010735e1b95ca724a2ff5))
* **linux:** :bug: use distinct device classes for laptop sensors ([c0f5fac](https://github.com/joshuar/go-hass-agent/commit/c0f5fac354e953111ab509d7c7e7627bfa2292e5))
* **linux:** :loud_sound: add repercussions of some settings being unavailable to warning messages ([af6fc62](https://github.com/joshuar/go-hass-agent/commit/af6fc62a1c9c9e4754c12680b56587d89e4ec3b0))


### Performance Improvements

* **agent:** :fire: remove unnecessary context creation ([80890aa](https://github.com/joshuar/go-hass-agent/commit/80890aa05f89bf0c192378ea37ee104be9262975))
* **dbusx:** :zap: more graceful dbus watch closure ([5724468](https://github.com/joshuar/go-hass-agent/commit/5724468dbb1ea422e8421c4ae914e6e4fd6f8f72))
* **hass:** :building_construction: remove sensor interfaces, use exported struct instead ([80c5780](https://github.com/joshuar/go-hass-agent/commit/80c57800d6c4d2bda691b0cccd63ee133dc9357b))
* **hass:** :fire: remove unnecessary context creation ([6dfd48a](https://github.com/joshuar/go-hass-agent/commit/6dfd48afc76c67ea8a10dbfb271cc29c8fe47ac6))


### Reverts

* **github:** :rewind: switch back to audit to check required access ([03b7e2a](https://github.com/joshuar/go-hass-agent/commit/03b7e2a6419943c25cea037274c03f28a5a75776))

## [10.2.1](https://github.com/joshuar/go-hass-agent/compare/v10.2.0...v10.2.1) (2024-09-15)


### Bug Fixes

* **linux:** :bug: correct power state tracking ([5bb5a4d](https://github.com/joshuar/go-hass-agent/commit/5bb5a4d9d11c8d174cda3eed9b4851b6a2345879))
* **linux:** :bug: correct units of cpufreq sensors ([21e7104](https://github.com/joshuar/go-hass-agent/commit/21e7104c5dac591f36978e5e0f46695f9e2a5fff))
* **linux:** :bug: make sure power controls pass required argument to D-Bus method call ([c44cd2c](https://github.com/joshuar/go-hass-agent/commit/c44cd2cefb0c3367d896a72db205d61324d562de))


### Performance Improvements

* **agent:** :building_construction: restructure preferences and hass client usage ([f6b0833](https://github.com/joshuar/go-hass-agent/commit/f6b08332dc8656b72101deb831a7ff19a565d3c0))
* **linux:** :zap: adapt prometheus trick for grabbing hwmon file data ([7b498bc](https://github.com/joshuar/go-hass-agent/commit/7b498bc71f22acc12b430543b66a0cce347b232b))

## [10.2.0](https://github.com/joshuar/go-hass-agent/compare/v10.1.1...v10.2.0) (2024-09-12)


### Features

* **hass:** :sparkles: add validation of sensor requests ([3e2c560](https://github.com/joshuar/go-hass-agent/commit/3e2c5602493c8910f6e14a63e3a095c969b5a8eb))
* **preferences:** :sparkles: add support for setting MQTT preferences via the command-line ([b49d0db](https://github.com/joshuar/go-hass-agent/commit/b49d0db5dc7305eedd37efef83e05ce72e3850eb))


### Bug Fixes

* **agent:** :bug: correct check on MQTT enabled for resetting agent ([8d57930](https://github.com/joshuar/go-hass-agent/commit/8d57930af86574dd3d053182865dc21ae04baaae))
* **agent:** :bug: re-add profiling webui support ([83e7c59](https://github.com/joshuar/go-hass-agent/commit/83e7c597ff3f8580c59cd2782ab64aa86338a596))
* **cli:** :bug: retain `--terminal` cli flag for "headless" mode ([e1f6f84](https://github.com/joshuar/go-hass-agent/commit/e1f6f8434646683d8983e09e397c998014f1100b))
* **cli:** :see_no_evil: ensure text files are included ([be5c04f](https://github.com/joshuar/go-hass-agent/commit/be5c04f4882f7f5c53d5fabcf7ae8f9d0c90a085))
* **container:** :bug: Alpine container fixes ([8317b8b](https://github.com/joshuar/go-hass-agent/commit/8317b8be249b6c891c0492d5a84a22eb72a94022))
* **dbusx:** :bug: introspect a method before calling to santize arguments ([9bbf0d9](https://github.com/joshuar/go-hass-agent/commit/9bbf0d964f88898a74d56fb8b62fd1eb85b2626f))
* **device:** :bug: more robust fetching of device values ([0be5b5f](https://github.com/joshuar/go-hass-agent/commit/0be5b5f6fa6ecacb266940c06bd8bf15232b762b))
* **hass:** :bug: actually retrieve and return response errors from HA ([f15d77e](https://github.com/joshuar/go-hass-agent/commit/f15d77e0d055d48b37d44cebed1a42d965e574f3))
* **hass:** :bug: not all sensors with a device type have units ([8d208b4](https://github.com/joshuar/go-hass-agent/commit/8d208b4a83854fa8ab694bf5cf0f8ad57fc75f15))
* **hass:** :mute: normal websocket closure should not warn (gws pkg update change) ([f1ce02d](https://github.com/joshuar/go-hass-agent/commit/f1ce02db47228189fe074dd564ce78b56701084b))
* **linux:** :bug: display at least some name if no display name was set for sensor ([96bd6a2](https://github.com/joshuar/go-hass-agent/commit/96bd6a221729c3a94e1f41371da61c707053e283))
* **linux:** :bug: don't return nil slice, return slice with len 0 ([1704d33](https://github.com/joshuar/go-hass-agent/commit/1704d3303736330186e8b4b53c1ee99515fb83fe))
* **linux:** :bug: ensure rate sensors have an initial value (of zero) for validation ([a3096d0](https://github.com/joshuar/go-hass-agent/commit/a3096d09f7916495174aeb749bd1c20ba932c9c6))
* **linux:** :bug: filter some uninteresting mountpoints from being disk usage sensors ([6ea66d7](https://github.com/joshuar/go-hass-agent/commit/6ea66d7d4606007658e21ca7dddf0e5641ac669c))
* **linux:** :bug: handle missing stats ([295a893](https://github.com/joshuar/go-hass-agent/commit/295a893897005d4f359107fb0f8312d99e3e3dbb))
* **linux:** :bug: only add values to context that are present/available ([27d49fc](https://github.com/joshuar/go-hass-agent/commit/27d49fcf7d8d15c5c38fab4425e2d9f1eed62fdb))
* **linux:** :zap: don't run problems worker if ABRT problems are not available in D-Bus ([fecb599](https://github.com/joshuar/go-hass-agent/commit/fecb599daf832d8b03ebce4410afd787d9502a40))
* **linux/hwmon:** :bug: fix naming of alarm sensors ([ee78240](https://github.com/joshuar/go-hass-agent/commit/ee782407a23e795282975b8af3ffeb16104720f0))
* **logging:** :zap: improve logging setup ([a3e05bb](https://github.com/joshuar/go-hass-agent/commit/a3e05bb036c5aad87f083138d0616353e40c43ca))
* **upgrade:** :bug: don't report an error if there is no need to upgrade ([ca3ba6e](https://github.com/joshuar/go-hass-agent/commit/ca3ba6e8561eb01b23d995f483bc9cc81b49fe97))
* **upgrade:** :bug: handle encountering nil when loading preferences ([52c2d64](https://github.com/joshuar/go-hass-agent/commit/52c2d64969f417fcb3e016b658ce22cc14ebbf62))


### Performance Improvements

* **agent:** :zap: handle signals with a context ([a719b92](https://github.com/joshuar/go-hass-agent/commit/a719b9212d4631a7cbbed723481998d4e88ba3e9))
* **agent:** :zap: improve protections against nil pointer exceptions ([303dc58](https://github.com/joshuar/go-hass-agent/commit/303dc58666d81ba554cc890337932888c0d00e6c))
* **commands:** :zap: improve protections against nil pointer exceptions ([be98b17](https://github.com/joshuar/go-hass-agent/commit/be98b17a00335939f6b845eebd69c62e9d6f137f))
* **linux:** :fire: remove unnecessary custom logger from mem worker ([2588436](https://github.com/joshuar/go-hass-agent/commit/25884363387d7815e52c5e81db9d8323fafbf6f0))
* **linux:** :recycle: store and fetch more values to/from context ([772fd56](https://github.com/joshuar/go-hass-agent/commit/772fd56489b78644ebe4d6102bb12f33575fee5e))
* **linux:** :zap: improve disk IO sensors ([6ef8dfb](https://github.com/joshuar/go-hass-agent/commit/6ef8dfb0184e468de8a5f5fabf6aeb936818f6b4))
* **linux:** :zap: improve protections against nil pointer exceptions ([2793806](https://github.com/joshuar/go-hass-agent/commit/27938062d47e0ab4ffe70d48aea64f1730987f87))
* **linux:** :zap: try to avoid dynamic sensor ID generation ([1013711](https://github.com/joshuar/go-hass-agent/commit/1013711a777de5a09f41c66360e17eb0fc8d2acf))
* **linux/hwmon:** :zap: rework hwmon sensors ([0164429](https://github.com/joshuar/go-hass-agent/commit/0164429ba85e5ace60c9f164c8655db35432a368))
* **linux/hwmon:** :zap: simplify sensor collection ([d145cab](https://github.com/joshuar/go-hass-agent/commit/d145cabb8292d7e332c8af18e4fa7e55a04b2519))
* **scripts:** :zap: improve protections against nil pointer exceptions ([191b7c0](https://github.com/joshuar/go-hass-agent/commit/191b7c017a8250af984b96aaf501adf1a4102383))

## [10.1.1](https://github.com/joshuar/go-hass-agent/compare/v10.1.0...v10.1.1) (2024-09-01)


### Bug Fixes

* **linux:** :bug: don't include mounts where stats cannot be retrieved for disk usage sensors ([82534c8](https://github.com/joshuar/go-hass-agent/commit/82534c8418ed5ebfbf0b96dfe63e9e497f9ae738))

## [10.1.0](https://github.com/joshuar/go-hass-agent/compare/v10.0.1...v10.1.0) (2024-08-27)


### Features

* **linux:** :recycle: rework memory usage sensors ([0930a5c](https://github.com/joshuar/go-hass-agent/commit/0930a5cc0770c81327c19ac0358f1e045cbc6805))
* **linux:** :sparkles: add support for alternative system partition mounts in cpu sensors ([55e6c78](https://github.com/joshuar/go-hass-agent/commit/55e6c783e982da3a44facf8c11635b390fd4a627))
* **linux:** :sparkles: improve cpufreq and introduce per core cpu usage sensors ([412fef1](https://github.com/joshuar/go-hass-agent/commit/412fef13c5136016d6b16d8fbb533ef60f05174a))
* **linux:** :sparkles: support ability to specify alternative mount points for system mounts via environment variables ([133142b](https://github.com/joshuar/go-hass-agent/commit/133142b3d94ad5db98a718ca7a9b80c68ea3acb9))


### Bug Fixes

* **linux:** :bug: ensure disk io stats are correct ([80f9a80](https://github.com/joshuar/go-hass-agent/commit/80f9a8011c7273856bde14e7dfc0ae9d9f0036a4))
* **linux:** :bug: ensure stats file is closed properly ([8c97527](https://github.com/joshuar/go-hass-agent/commit/8c975277bc7d4dac7ce1496770c2639172652dc3))
* **linux:** :bug: fix bootime value after recent changes ([fa977eb](https://github.com/joshuar/go-hass-agent/commit/fa977eb024b0bba5911248f62b03a533fdb3248a))
* **linux:** :bug: usage count values should be ints not strings ([af01217](https://github.com/joshuar/go-hass-agent/commit/af01217d4847e8795f0d9a783358308edac3d5ec))
* **linux:** :mute: don't report problems fetching hardware sensors at default log level ([d271fee](https://github.com/joshuar/go-hass-agent/commit/d271feebfbf95df196cbbef3f4b74cc19408a55d))
* **linux:** :recycle: rework time, io and cpu sensors ([01593c7](https://github.com/joshuar/go-hass-agent/commit/01593c7d2f2b5691ee919848c0f9771510d095a9))


### Performance Improvements

* **linux:** :recycle: improve disk usage sensors ([712309c](https://github.com/joshuar/go-hass-agent/commit/712309c77df8ef8a01f9a5b38a632db5ed609144))
* **linux:** :zap: fetch load averages internally ([067e2ca](https://github.com/joshuar/go-hass-agent/commit/067e2caef0f8e33fcd93f85af9381684041b045d))
* **linux:** :zap: increase interval between polling for CPU frequency measurements ([c44eeea](https://github.com/joshuar/go-hass-agent/commit/c44eeea2c56403d36f7941bbd648b7613c3be97b))
* **linux:** :zap: optimise disk io sensors ([2124a77](https://github.com/joshuar/go-hass-agent/commit/2124a778c05f4b2a11c959cb20474da594c058fc))

## [10.0.1](https://github.com/joshuar/go-hass-agent/compare/v10.0.0...v10.0.1) (2024-08-21)


### Bug Fixes

* **agent,hass:** :bug: fix registration flow after hass client refactoring ([5e4a9ba](https://github.com/joshuar/go-hass-agent/commit/5e4a9ba238ea3dd370610086b6fc719ecae93ac4))
* **hass:** :bug: pass required request type to NewRequest ([ec0a7e8](https://github.com/joshuar/go-hass-agent/commit/ec0a7e8db76bb04ec7b40a1fdd3c5a27de6ddd91))
* **linux:** :bug: ensure D-Bus command topic is unique ([d403dbb](https://github.com/joshuar/go-hass-agent/commit/d403dbbe3867764cc43e014189b0ee4d5ad27f3e))
* **linux:** :bug: make sure MQTT topics are unique for power and session controls ([e77b927](https://github.com/joshuar/go-hass-agent/commit/e77b9274aac249d6e1ddbc61457ed441c19c6a6c))
* **linux:** :mute: reduce log spam if a mountpoint usage cannot be retrieved ([f4aea29](https://github.com/joshuar/go-hass-agent/commit/f4aea2963a616f6b10593c9df8ea3dbb692f8e79))


### Performance Improvements

* **agent,hass:** :zap: refactor sensor processing ([c7c3ff3](https://github.com/joshuar/go-hass-agent/commit/c7c3ff383b529051ebfe1608caee38eb8a25f46e))

## [10.0.0](https://github.com/joshuar/go-hass-agent/compare/v9.6.0...v10.0.0) (2024-08-17)


### ⚠ BREAKING CHANGES

* **agent:** the device representing Go Hass Agent in Home Assistant has been renamed from the generic "Go Hass Agent" to the hostname of the device running Go Hass Agent.
* **prefs:** The default app id has changed, which changes the path to the agent configuration. As such, the agent will need to be re-registered with Home Assistant.
* major internal update
* MQTT preferences have been renamed in the config file. They now sit under their own heading. Existing MQTT preferences are not migrated to the new settings.

### Features

* :sparkles: add an upgrade command to help with upgrading after major release ([d244aa9](https://github.com/joshuar/go-hass-agent/commit/d244aa91cd502a52823377cd22203e91b760e90e))
* **agent:** :sparkles: add support for number controls with custom MQTT commands ([09d44c4](https://github.com/joshuar/go-hass-agent/commit/09d44c41737f25261562c7150b379300215bd7cb))
* **agent:** :sparkles: use a nicer name for the app "ID" that is exposed by Fyne to the desktop environment ([87d9ae4](https://github.com/joshuar/go-hass-agent/commit/87d9ae4ba97a0e50c50b7a3a923fe610f201a6f9))
* **agent:** rename the MQTT device ([da65683](https://github.com/joshuar/go-hass-agent/commit/da656837ea3bd1e905f37b8ce8df1ec5b2148804))
* **dbusx:** :sparkles: add support for watching on arg namespace ([22bc528](https://github.com/joshuar/go-hass-agent/commit/22bc528cb6abc687ccc55dda41fc86f91081b2b4))
* **linux:** :sparkles: add basic webcam view/control ([2c30336](https://github.com/joshuar/go-hass-agent/commit/2c30336662f49a0d031c9a392e2360542df40fb7))
* **linux:** :sparkles: add CPU frequency sensors ([6b7b91f](https://github.com/joshuar/go-hass-agent/commit/6b7b91fd0229bac4c5129e492e0359035e910f57))
* **linux:** :sparkles: add sensor tracking media status of any MPRIS compatibile player on the system ([5915521](https://github.com/joshuar/go-hass-agent/commit/5915521433e91dd7104cadde0fda2ffb72755245))
* **linux:** :sparkles: better screen/session controls ([17759e2](https://github.com/joshuar/go-hass-agent/commit/17759e2c21f64d51772a032494fa9bbb432ea877))
* **linux:** :zap: improve active/running apps sensor code ([2971035](https://github.com/joshuar/go-hass-agent/commit/29710355aa50e16da7c2fea270ce8e310d407ea9))
* **linux:** :zap: increase polling (frequency) of cpu usage (%) sensor updates ([8d4c9da](https://github.com/joshuar/go-hass-agent/commit/8d4c9da70dbb969cf2ca22d77a24ae8265c9f1cd))
* **preferences:** :loud_sound: improve messages shown when preferences are not valid ([cf4dd0a](https://github.com/joshuar/go-hass-agent/commit/cf4dd0a8fd9d1e9cf6d07b784d4e6e724fd6c154))


### Bug Fixes

* **agent:** :bug: censure app id is set correctly in agent ([c126d45](https://github.com/joshuar/go-hass-agent/commit/c126d455268f5fa4d3e95498cb2c2560543d7f88))
* **agent:** :bug: correct mock name so go generate doesn't crash ([e7ac7fb](https://github.com/joshuar/go-hass-agent/commit/e7ac7fb779ac5a50e1b29c7d8c24dd0408dae49c))
* **agent:** :bug: don't exit MQTT runner if MQTT commands cannot be set up ([2417c4c](https://github.com/joshuar/go-hass-agent/commit/2417c4c4d0113025d7baaa4c83b964bccecd0e8f))
* **agent:** :bug: don't run MQTT workers if there are no workers ([bcc7636](https://github.com/joshuar/go-hass-agent/commit/bcc76363fc04c9cef248e8a5fa7173d80c71cf50))
* **agent:** :bug: get HA config needs rest API URL ([a549d39](https://github.com/joshuar/go-hass-agent/commit/a549d39db076d5d2ca257f57b71bd973de000c00))
* **agent:** :bug: support passing registration parameters via command-line when running in graphical mode ([4e35159](https://github.com/joshuar/go-hass-agent/commit/4e351591220bf21ef94691688388004b3fdf1aab))
* **agent:** :bug: sync sensor disabled state between registry and Home Assistant ([cc8d89c](https://github.com/joshuar/go-hass-agent/commit/cc8d89c544a7c0e452e1541f58f5082fe6f46cb5))
* **agent:** :loud_sound: write a log message when agent is registered ([1cc6168](https://github.com/joshuar/go-hass-agent/commit/1cc61689464214f39efe11eb635765254965cd32))
* **agent:** :mute: fix logging when no MQTT commands are defined ([718b0b0](https://github.com/joshuar/go-hass-agent/commit/718b0b0fac9e0d33b887e20bffba7549e5602921))
* **linux:** :bug: only provide power controls that are available on the device ([540559d](https://github.com/joshuar/go-hass-agent/commit/540559d84a612bc36220809d7826787d9a61e736))
* **logging:** :bug: (again again) fix create directory for logfile ([b73faff](https://github.com/joshuar/go-hass-agent/commit/b73faffe6832462777de76918784908cd43d3f2c))
* **logging:** :bug: (again) create directory for log file if not exists ([e60a7f0](https://github.com/joshuar/go-hass-agent/commit/e60a7f0c0e6c88efeb62d3976ca7876127dafcba))
* **logging:** :bug: handle a non-existent directory for the log file (auto-create if necessary) ([03eb8e4](https://github.com/joshuar/go-hass-agent/commit/03eb8e48c25b8dacc034d3f8376d139111faa260))
* **logging:** :loud_sound: don't crash if we can't write to the log file ([f92cf2f](https://github.com/joshuar/go-hass-agent/commit/f92cf2f39be1f3967b2ea920b0eb7ab4d7d9890b))
* **logging:** :loud_sound: fix log file path and level details ([1bf85ea](https://github.com/joshuar/go-hass-agent/commit/1bf85eaa2581490347f3a594108ea03c857fcdcb))
* **scripts:** :bug: don't return an open channel that will never close if there are no scripts ([d17d4e7](https://github.com/joshuar/go-hass-agent/commit/d17d4e7aa634554dec8fb742325fcd9c9f60faa1))
* **ui:** :bug: re-add default server to list of servers when registering agent ([c9cfd9c](https://github.com/joshuar/go-hass-agent/commit/c9cfd9c1a83f6a9d875f8225fb86486d5fa2ac4e))


### Performance Improvements

* **agent:** :fire: remove unnecessary goroutines and waitgroups ([0f01468](https://github.com/joshuar/go-hass-agent/commit/0f0146882bd6ce6aa954580161e31ee2fb000778))


### Code Refactoring

* improved MQTT functionality ([766fcce](https://github.com/joshuar/go-hass-agent/commit/766fcce98f7cd1f24aadbe85fb1d371da401ffde))
* major internal update ([61926f6](https://github.com/joshuar/go-hass-agent/commit/61926f6be1faf734144d4a75e0b16d7193fca7c5))
* **prefs:** change default app id ([6125a45](https://github.com/joshuar/go-hass-agent/commit/6125a45070e1ced488e6d9e39102c0354d2a0793))

## [9.6.0](https://github.com/joshuar/go-hass-agent/compare/v9.5.2...v9.6.0) (2024-07-27)


### Features

* :loud_sound: improve logging output ([5f12810](https://github.com/joshuar/go-hass-agent/commit/5f12810f2331fd6cf8ba8506d0bc2b78231220e0))
* **agent:** :loud_sound: improve agent logging ([a35fcb2](https://github.com/joshuar/go-hass-agent/commit/a35fcb26f3afd819c0dd0160ee71a183ca6af4c6))
* **linux:** :zap: D-Bus overhaul ([2cf7dd4](https://github.com/joshuar/go-hass-agent/commit/2cf7dd47337140065424409b0e275acee1705d58))


### Bug Fixes

* **agent:** :bug: actually save agent preferences and show better dialogs for success/fail ([dfd1c47](https://github.com/joshuar/go-hass-agent/commit/dfd1c478b1106d9d984aec34c11f3a8335c902f4))
* **agent:** :bug: make sure cron scheduler is stopped gracefully ([d631411](https://github.com/joshuar/go-hass-agent/commit/d631411283da9616950a9f0d42b3ea5a0f81538e))
* **hass:** :loud_sound: make request body more readable ([bf1f6c4](https://github.com/joshuar/go-hass-agent/commit/bf1f6c4b860d4806951aa24582fcbc0343150ba0))
* **linux:** :loud_sound: correct worker type in error message ([5a62443](https://github.com/joshuar/go-hass-agent/commit/5a624432ecde9c1c1381a7c4f665b899a5b4d778))


### Performance Improvements

* :zap: improve ability to stop and start sensor workers ([ad045c6](https://github.com/joshuar/go-hass-agent/commit/ad045c6aa550666a350c0fcc7f04ddaa7afab99d))


### Reverts

* **go:** :rewind: go back to previous go generate incantation ([2495017](https://github.com/joshuar/go-hass-agent/commit/2495017ad1729aba94ae069915e5ed785d63402f))

## [9.5.2](https://github.com/joshuar/go-hass-agent/compare/v9.5.1...v9.5.2) (2024-07-07)


### Bug Fixes

* **agent:** :bug: don't hang on register command if already registered ([37e29cc](https://github.com/joshuar/go-hass-agent/commit/37e29cc14fa3458fcfdb2872fc99d0c9729bd2a8))
* **agent:** :bug: ensure preferences are set in context *after* registration is completed ([96bf97f](https://github.com/joshuar/go-hass-agent/commit/96bf97f7e4c211547a66400e22c800534ad65e68))
* **agent:** :bug: make sure agent quits if registration process is cancelled ([e5acc53](https://github.com/joshuar/go-hass-agent/commit/e5acc533aeda670bdd2416a8145d6e22d43d2365))
* **hass:** :bug: don't add State/Device classes or Sensor Type values to responses if they are not set ([022c80f](https://github.com/joshuar/go-hass-agent/commit/022c80f239154d293cf8631fc2ecdd9e3d58bd63))
* **linux:** :bug: handle quoted and unquoted values in /etc/os-release correctly ([bdf4fce](https://github.com/joshuar/go-hass-agent/commit/bdf4fcef03ae813f3d751926f74090795d8e34c4))
* **mage:** :bug: correct invocation for ldflags for fyne-cross ([4ff7801](https://github.com/joshuar/go-hass-agent/commit/4ff7801bb7010d4e4b757af303caab227a4a830c))


### Performance Improvements

* :recycle: improve Home Assistant API request handling ([36aee1c](https://github.com/joshuar/go-hass-agent/commit/36aee1c1ba92be06a85834fdc8e35c94545a5250))
* :sparkles: preferences rewrite ([c15f486](https://github.com/joshuar/go-hass-agent/commit/c15f486fd5eecbbc99b68fb7cf1fb1758f4b8177))
* **hass:** :recycle: rework request logic ([2031c88](https://github.com/joshuar/go-hass-agent/commit/2031c888fdd8e1dcae7765dd832f38a0ddc30270))

## [9.5.1](https://github.com/joshuar/go-hass-agent/compare/v9.5.0...v9.5.1) (2024-07-02)


### Bug Fixes

* **linux:** :bug: don't try to create controls if they are unavailable ([f2fafbc](https://github.com/joshuar/go-hass-agent/commit/f2fafbc85b2e39d134ea570059ce588a784cfde4))
* **scripts:** :bug: improve error and argument handling ([a37f20c](https://github.com/joshuar/go-hass-agent/commit/a37f20cf1fa7486dc859f40d82ad39fc7304f983))


### Performance Improvements

* **agent:** :fire: remove unneeded and complicated koanf usage ([b0740e8](https://github.com/joshuar/go-hass-agent/commit/b0740e88aeca48b7d7f7d6762219271d25961064))

## [9.5.0](https://github.com/joshuar/go-hass-agent/compare/v9.4.0...v9.5.0) (2024-06-27)


### Features

* **agent:** :sparkles: support "switch" type custom MQTT commands ([26f3272](https://github.com/joshuar/go-hass-agent/commit/26f3272eb1d746cd30f28918afd03b1a2c880334))
* **container:** :sparkles: support cross-compilation for container images ([37489ec](https://github.com/joshuar/go-hass-agent/commit/37489ec4beb3a7d69b0eedb0c904e023e3223e6f))


### Bug Fixes

* **hass:** :bug: ensure sensor attributes are correctly marshaled ([ec6fe29](https://github.com/joshuar/go-hass-agent/commit/ec6fe29c02071073f05b54eaee4dbe56ffd830bc))
* pass correct arch to apt-get ([e536162](https://github.com/joshuar/go-hass-agent/commit/e53616239da58d3104e06d0bd99ea0aca34cd951))

## [9.4.0](https://github.com/joshuar/go-hass-agent/compare/v9.3.0...v9.4.0) (2024-06-17)


### Features

* :sparkles: add a framework for allowing sensor workers to be start/stopped ([988626a](https://github.com/joshuar/go-hass-agent/commit/988626a8d7c3f8b44bbd6a37db0a863ee2962531))
* **agent:** :sparkles: add framework for running user-defined commands via MQTT buttons/switches ([61ff2af](https://github.com/joshuar/go-hass-agent/commit/61ff2af256b4a639e07f80824655c261c4034f5f))
* **agent:** :sparkles: support ignore URLs registration option on command-line ([17927ed](https://github.com/joshuar/go-hass-agent/commit/17927ed5c69538b04483339c7df50a3bd73daa5b))


### Bug Fixes

* :bug: make log-level argument an enum ([abee169](https://github.com/joshuar/go-hass-agent/commit/abee169d53738d88caa40f7eb00b437cf7846b3a))
* :bug: pass a context to worker creation functions ([aeca410](https://github.com/joshuar/go-hass-agent/commit/aeca41062fa3db781ccc241bf272bd34640dd6a0))
* **agent:** :art: mqtt commands clean-up ([56c55d2](https://github.com/joshuar/go-hass-agent/commit/56c55d295d4c7e640b43192416639cdabd5ba039))
* **agent:** :bug: don't bail on registration if no preferences file was found but default preferences were returned ([74b52f1](https://github.com/joshuar/go-hass-agent/commit/74b52f1a5c0af83618d883853b97d0f840831ef1))
* **agent:** :bug: ensure sensor workers can remain running when running headless ([4fd534b](https://github.com/joshuar/go-hass-agent/commit/4fd534b9133d979527fa85b445ec886707c3f88c))
* **agent:** :bug: hide worker failure log message behind debug logging as such failures are non-critical ([3c4a4c5](https://github.com/joshuar/go-hass-agent/commit/3c4a4c52c609b6a0597f6c7b8e6b1547285a0e2f))
* **hass:** :speech_balloon: show a more informative error message when registration inputs validation fails ([57876e4](https://github.com/joshuar/go-hass-agent/commit/57876e46e14ab20fa496e40538f3d2b106c9d72f))
* **linux:** :sparkles: don't provide a location sensor when not running on a laptop device ([acac8d7](https://github.com/joshuar/go-hass-agent/commit/acac8d74e14f216db3d2ebfe6297114e0365528e))
* **preferences:** :bug: correctly detect default preferences as not a fatal error ([3297099](https://github.com/joshuar/go-hass-agent/commit/329709911c667d7a55c4d095576f1fd9f99d4613))
* **translations:** :art: ensure translator uses appropriate language ([c0cd489](https://github.com/joshuar/go-hass-agent/commit/c0cd489bcf42de4f5ab22ddac574e6deb0292e10))

## [9.3.0](https://github.com/joshuar/go-hass-agent/compare/v9.2.0...v9.3.0) (2024-06-01)


### Features

* **ui:** :sparkles: use a new icon and clean up text in UI ([01bd2f3](https://github.com/joshuar/go-hass-agent/commit/01bd2f30fe48be59959a9a5108cdebf675eae6a7))


### Bug Fixes

* **linux:** :bug: actually report correct distribution, distribution and kernel version as sensors ([6db5082](https://github.com/joshuar/go-hass-agent/commit/6db5082472f2d528ca30bb65ee58e777064a360c))
* **ui:** :bug: embed correct icon ([6f9f412](https://github.com/joshuar/go-hass-agent/commit/6f9f412d456ac87c976a16e7d5e0859f08d0986b))

## [9.2.0](https://github.com/joshuar/go-hass-agent/compare/v9.1.1...v9.2.0) (2024-05-23)


### Features

* **dbusx:** :art: improved new WatchBus function with more support for native D-Bus match types ([6066daa](https://github.com/joshuar/go-hass-agent/commit/6066daaf673a67d01c9def418b0cb592477da4b8))
* **linux:** :sparkles: add desktop session idle sensor ([1bd5d08](https://github.com/joshuar/go-hass-agent/commit/1bd5d08098d58bc1174a3264445ae3bcbf3cb921))
* **linux:** :sparkles: migrate to different pulseaudio library ([e5c576c](https://github.com/joshuar/go-hass-agent/commit/e5c576cbe2c5161082be33ff8834ac15f92bf0b4))


### Bug Fixes

* **agent:** :arrow_up: update go-hass-anything to fix authentication issues with MQTT ([51d5f54](https://github.com/joshuar/go-hass-agent/commit/51d5f54db4076da97ffd56199f4e0d4b97ec07ad))
* **dbusx:** :bug: better protection against nil pointer exception in bus connection from context retrieval ([8ddf562](https://github.com/joshuar/go-hass-agent/commit/8ddf56247e6ee9fc7b9aedf9a2db48ac050d0e06))
* **linux:** :bug: request correct idleTime property ([312179f](https://github.com/joshuar/go-hass-agent/commit/312179f50a50c410f4857dda9d7b2ef98e5b652f))
* **linux:** :fire: remove spew ([d454883](https://github.com/joshuar/go-hass-agent/commit/d45488394b84e3a8d11e3a212013516f08eca4da))
* **linux:** :label: add idle sensor type ([36c8319](https://github.com/joshuar/go-hass-agent/commit/36c83192dbf803ec319025d87089ff1f3546409c))
* **ui:** :lipstick: Fyne -&gt; Fyne Settings ([42b8c76](https://github.com/joshuar/go-hass-agent/commit/42b8c769e31cd7721642acfa862b31d677ef8e1e))


### Performance Improvements

* **dbusx:** :sparkles: support checking if multiple properties/signals have changed in a single watch ([473a1ba](https://github.com/joshuar/go-hass-agent/commit/473a1ba6ffb585409b7f1d2d0d817cd48f8185e2))
* **linux:** :zap: all power sensors using D-Bus use new dbusx.WatchBus function ([03ab63e](https://github.com/joshuar/go-hass-agent/commit/03ab63e6f7fca3b1f3dc0886486e7c7e7cc5edde))
* **linux:** :zap: complete migration of battery sensor code to dbusx.WatchBus ([0055e56](https://github.com/joshuar/go-hass-agent/commit/0055e5653009d548be181109758123520b9582c6))
* **linux:** :zap: rework battery sensor to use new dbusx.WatchBus function ([77b172a](https://github.com/joshuar/go-hass-agent/commit/77b172a55cf396b567eb99c9175eef71eed40d83))
* **linux:** :zap: rework desktop preferences sensors to use dbusx.WatchBus function ([46f6aaa](https://github.com/joshuar/go-hass-agent/commit/46f6aaae858dfbcddc335883a3683db0303b1f32))
* **linux:** :zap: rework laptop sensors to use dbusx.WatchBus ([942d089](https://github.com/joshuar/go-hass-agent/commit/942d089318bcd405dde4b12391ae92c1796435a8))
* **linux:** :zap: rework network sensors to use new dbusx.WatchBus function ([485a1e6](https://github.com/joshuar/go-hass-agent/commit/485a1e6c2560eda64bc0a49a4d1e993a1cd5d463))
* **linux:** :zap: rework wifi properties sensors to use dbusx.WatchBus function ([d5189ce](https://github.com/joshuar/go-hass-agent/commit/d5189cec086e9bf085526ea192229ba8b0c1c5e7))
* **linux:** :zap: use new D-Bus watch method for location updates ([a441579](https://github.com/joshuar/go-hass-agent/commit/a441579e345cf12f48508cbf388350e28f4e83b8))
* **linux:** :zap: use new D-Bus watch method for power state updates ([dabed9f](https://github.com/joshuar/go-hass-agent/commit/dabed9f7b148157d8ea53e570bd4cbdbb1577567))
* **linux:** :zap: user session tracking sensor use new dbusx.WatchBus function ([4b7adc2](https://github.com/joshuar/go-hass-agent/commit/4b7adc23b5ebc270a5d86f34b2add8e30cbdf352))

## [9.1.1](https://github.com/joshuar/go-hass-agent/compare/v9.1.0...v9.1.1) (2024-05-18)


### Bug Fixes

* **agent:** :ambulance: fix unavailable location sensor causing crashes ([a9a54ac](https://github.com/joshuar/go-hass-agent/commit/a9a54ac324ce5490a788dcfcaa93e0c7f3292262))

## [9.1.0](https://github.com/joshuar/go-hass-agent/compare/v9.0.0...v9.1.0) (2024-05-18)


### Features

* **agent:** :building_construction: start creating the framework for more efficient sensor updates ([2b98a89](https://github.com/joshuar/go-hass-agent/commit/2b98a893c04fa711248bd0cf7dbb4170f88d822b))
* **agent:** :sparkles: initial migration from cobra to kong ([e93f733](https://github.com/joshuar/go-hass-agent/commit/e93f7335056b0ebb445b92dd3764b3ee472a54f8))
* **dbusx:** :sparkles: new function for simpler watch creation ([91cd040](https://github.com/joshuar/go-hass-agent/commit/91cd040771c0983c14510101a26264a118f66c45))
* **linux:** :sparkles: detect machine chassis type ([c5ac91a](https://github.com/joshuar/go-hass-agent/commit/c5ac91ae5e3f241c19fb32d938d18bad9d19849d))
* **linux:** :sparkles: new improved laptop sensors for lid closed state, docked state and external power connected state ([c6ca2b6](https://github.com/joshuar/go-hass-agent/commit/c6ca2b6ed3a8ab096fa7ace571e0880ed963dd19))


### Bug Fixes

* **agent:** :bug: enable more profiling options ([0717f92](https://github.com/joshuar/go-hass-agent/commit/0717f92f8d8ad7a6d5327230ba6cf0e364dc93d4))
* **agent:** :bug: ensure a default app ID is set when not specified ([e5c7265](https://github.com/joshuar/go-hass-agent/commit/e5c7265be4ee3e149fbd01dc78fd95988544338c))
* **agent:** :bug: ensure log level is set appropriately on start ([a00c742](https://github.com/joshuar/go-hass-agent/commit/a00c742da99cf7a19054c0ded51d94219716cf45))
* **agent:** :bug: if no GUI detected, default to running headless (and show a warning) ([6f9469d](https://github.com/joshuar/go-hass-agent/commit/6f9469dd0721da5a0f5bdc871ff27268b49ca69f))
* **agent:** :bug: if we cannot fetch the Home Assistant config, don't display its details on about window ([2241450](https://github.com/joshuar/go-hass-agent/commit/22414503e01f1ffae094c45855b155623dd2fab5))
* **agent:** :bug: return nil if cannot fetch Home Assistant config ([e731aa2](https://github.com/joshuar/go-hass-agent/commit/e731aa2dd4316ea791e06e54e7c71fbf724332fd))
* **agent:** :lipstick: merge the preferences submenu back into the main menu of the tray icon ([097c35e](https://github.com/joshuar/go-hass-agent/commit/097c35e482e9b837622f397653e001439028178d))
* **agent:** :zap: better handling of sensor channels ([f3724df](https://github.com/joshuar/go-hass-agent/commit/f3724dfae7f5daee8fe27fc240e0d6cfd18bbe83))
* **container:** :bug: use correct run command in container ([2f42ee8](https://github.com/joshuar/go-hass-agent/commit/2f42ee8691560beca7f69271c28027ff97a73216))
* **container:** :sparkles: better container defaults (and docs updates to match) ([e9a3cb7](https://github.com/joshuar/go-hass-agent/commit/e9a3cb7e897d7fd1c1fd8eff9d4ea5e27231ad19))
* **hass:** :bug: catch potential nil panic from error condition and handle appropriately ([a7a6dec](https://github.com/joshuar/go-hass-agent/commit/a7a6dec5c6e0febd5251910e804c820f3fdb76ba))
* **linux:** :bug: avoid sending on closed channel ([cbf95fc](https://github.com/joshuar/go-hass-agent/commit/cbf95fcc549529735202ba6ea622a4a178acfc20))
* **linux:** :bug: better problematic battery handling to avoid nil panics ([cda8403](https://github.com/joshuar/go-hass-agent/commit/cda84037227025d18caa8e99c0128ba3045482c9))
* **linux:** :bug: detect and warn on unrecognised wifi properties ([6e12013](https://github.com/joshuar/go-hass-agent/commit/6e12013117801c1c075896683e069188ed3713df))
* **linux:** :bug: keep device id constant for MQTT ([fcea433](https://github.com/joshuar/go-hass-agent/commit/fcea43394012131894b3b9c24a4788e9d84037ff))
* **linux:** :bug: power profile sensor should work again ([73f71d2](https://github.com/joshuar/go-hass-agent/commit/73f71d2169623a767695f6ae6a7b7f871a9f9cd3))
* **linux:** :bug: shutdown connection state monitor gracefully ([db2d165](https://github.com/joshuar/go-hass-agent/commit/db2d165cf6ce82e8fcec21661923923ac3966e35))
* **linux:** :bug: source distro information from /etc/os-release for registration ([a8d76c6](https://github.com/joshuar/go-hass-agent/commit/a8d76c6a177837b7641db6d5601902a0efd0ccb4))
* **linux:** :fire: remove deprecated channel close ([ab277af](https://github.com/joshuar/go-hass-agent/commit/ab277af063aa3e154e3fb60d8d48da3c026a3839))
* **linux:** :fire: remove unreleased idle sensor updater ([28d23e3](https://github.com/joshuar/go-hass-agent/commit/28d23e3338226648787a2622ea4e5a7a9b23e4ae))
* **linux:** :lock: as per recommendations, don't use the actual device id, generate a random one ([5d69005](https://github.com/joshuar/go-hass-agent/commit/5d690059415b404f870f6822700283363ac0352e))
* **linux:** :zap: detect if we have a lid, don't bother monitoring if we don't ([61c64b6](https://github.com/joshuar/go-hass-agent/commit/61c64b66d7bc65e99563d429a7a29247e9781256))
* **linux:** :zap: gracefully close pulseaudio connection on shutdown ([8df123e](https://github.com/joshuar/go-hass-agent/commit/8df123eb697f32941b8966d115be807b2d5d2acc))
* **linux:** :zap: gracefully shutdown mqtt message channel ([8d06d9e](https://github.com/joshuar/go-hass-agent/commit/8d06d9ecc518dc6fb850380fa115ecca4afd3710))
* **linux/proc:** :bug: handle unable to split /proc/diskstats ([7300ac9](https://github.com/joshuar/go-hass-agent/commit/7300ac9a05d2d17e68028fd2616b34f7f8e94547))


### Performance Improvements

* **linux:** :zap: send cpu load/usage state on start ([a1ea881](https://github.com/joshuar/go-hass-agent/commit/a1ea881daf51578b814204896b485bbec0a41c32))
* **linux:** :zap: utilise new D-Bus helper function for watching app state ([fe41e91](https://github.com/joshuar/go-hass-agent/commit/fe41e918fb12fbea7ad669c539559e4ff537bd46))

## [9.0.0](https://github.com/joshuar/go-hass-agent/compare/v8.0.0...v9.0.0) (2024-05-07)


### ⚠ BREAKING CHANGES

* **linux:** This commit changes the disk IO sensors so they are sourced from SysFS. This allows better filtering of devices to avoid creating sensors for partitions and other virtual devices. There should only be sensors created for physical disks, mdadm disks and device-mapper disks. Where possible, an attribute is available containing the model name of the disk. Additionally, a sensor will be created for total read/write counts and rates for all physical disks (excluding mdadm/device-mapper). Some entity IDs may have changed so be sure to check automations and dashboards and adjust as necessary.
* **linux/hwmon:** This is another refactoring of the hardware sensor (hwmon) parsing code. This hsould handle duplicate devices and generate unique sensors for all of them. As a result entities in Home Assistant will be renamed (again) which may break any automations and other functionality using the current names.
* This commit is a major refactoring of the MQTT functionality coinciding with changes to the underlying library that powers it. **MQTT entities have been renamed, which will result in some breakage of automations and features in Home Assistant.** No functionality has been lost however, and this change will make it easier to add additional controls and features powered by MQTT to Go Hass Agent.

### Features

* **agent:** :sparkles: add a `--no-log-file` command-line option to not write a log file ([59f2ce5](https://github.com/joshuar/go-hass-agent/commit/59f2ce53a16c023ab4eabc4df329ab109b4b671c))
* **linux:** :sparkles: add a volume mute switch ([3b8eb54](https://github.com/joshuar/go-hass-agent/commit/3b8eb54fcab8a80c1bfaa9f3d4443164b230e988))
* **linux:** :sparkles: add volume level control ([cac7077](https://github.com/joshuar/go-hass-agent/commit/cac70771e118250fb7fbbe8ba633e278a52288b7))
* **linux:** :sparkles: don't send swap sensors if there is no swap enabled ([f8508e2](https://github.com/joshuar/go-hass-agent/commit/f8508e209a863b7de67cae4b29260ca9354ed962))
* **linux:** improved disk IO sensors ([179f94e](https://github.com/joshuar/go-hass-agent/commit/179f94ebc3065e67d7422847f29d33ae8b7bd79d))
* **preferences:** :sparkles: add a function to return MQTT origin info for use in MQTT code ([2ab73bb](https://github.com/joshuar/go-hass-agent/commit/2ab73bba80a52eee9c6d869c179329172e72c3d0))


### Bug Fixes

* **agent:** :fire: remove testing commands ([6bd339d](https://github.com/joshuar/go-hass-agent/commit/6bd339def5b370d5771f068bed55e8a6e3d57609))
* **linux:** :bug: correct string version of disk rate units ([85a4164](https://github.com/joshuar/go-hass-agent/commit/85a41641f327d0c0000f0f317feea7aae74c197b))
* **linux:** :bug: disk reads/writes sensors should not use data size device class ([46f47bd](https://github.com/joshuar/go-hass-agent/commit/46f47bdedcd6f2a7bedca4d51c8d42d75dcf75c6))
* **linux:** :bug: don't block sending initial power state sensor update ([5e2100a](https://github.com/joshuar/go-hass-agent/commit/5e2100abd3246b679616657ecf6d01cea198144d))
* **linux:** :bug: don't block sending user sensor updates ([d6a982d](https://github.com/joshuar/go-hass-agent/commit/d6a982dd4091f23206e4ac7a7353886bd0ed9d7c))
* **linux:** :bug: ensure disk read/write count sensors have correct units ([da4f805](https://github.com/joshuar/go-hass-agent/commit/da4f805e61b15481661cdb78ba6b4ae110b36da5))
* **linux:** :bug: fix broken D-Bus control ([a49fce1](https://github.com/joshuar/go-hass-agent/commit/a49fce14b83079e6e06efebfadc65b7c0d9fae73))
* **linux:** :fire: remove spews (debugging) ([5e6aeb6](https://github.com/joshuar/go-hass-agent/commit/5e6aeb61ddd7a798e2ee2a268162b6452f9a2a7d))
* **linux:** :zap: ensure sending version sensors doesn't block ([838fd1e](https://github.com/joshuar/go-hass-agent/commit/838fd1e82bf440bf04225268350c3e0be5cdb488))
* **linux:** :zap: use unbuffered channel for disk usage sensors ([9b8b50a](https://github.com/joshuar/go-hass-agent/commit/9b8b50a5d34a71cbc2519c81b0e4e36d310edd8b))
* **linux:** :zap: use unbuffered channel for hardware sensor updates ([d4bbee2](https://github.com/joshuar/go-hass-agent/commit/d4bbee22fcd4262e1524997869789fdc8dcf0f50))
* **linux:** :zap: use unbuffered channel for time sensors ([13bf514](https://github.com/joshuar/go-hass-agent/commit/13bf51409ca27395bee476c9e64251c64c7b36d0))
* **linux/hwmon:** refactor sensor parsing (again) ([de865f1](https://github.com/joshuar/go-hass-agent/commit/de865f1ee8f89117581875796deb78efb585b9a8))
* **linux/proc:** :bug: fix stringer generation ([b0e5dc8](https://github.com/joshuar/go-hass-agent/commit/b0e5dc82d8362993aa2c986de195c8f9493839b5))


### Code Refactoring

* major MQTT functionality refactor ([decd825](https://github.com/joshuar/go-hass-agent/commit/decd825a6b07897513a12bebbaa188ad1746620b))

## [8.0.0](https://github.com/joshuar/go-hass-agent/compare/v7.3.1...v8.0.0) (2024-04-27)


### ⚠ BREAKING CHANGES

* **linux:** When more than one chip exposed by the hwmon userspace API have the same name, the agent was not treating them as unique. This commit fixes the code to ensure every chip gets its own sensors. This unfortunately required changing the naming format of all chips, so will result in new sensors being recorded in Home Assistant.

### Features

* **device:** :sparkles: add an agent version sensor ([92be2e1](https://github.com/joshuar/go-hass-agent/commit/92be2e1392354096b9e892827a5c8a36ed32fb99))
* **linux:** :sparkles: add Linux device IO rate sensors ([1011ea3](https://github.com/joshuar/go-hass-agent/commit/1011ea368fe78fa42b36a5c3de556780faef55d3))
* **linux:** :sparkles: add sensors for accent color and color scheme type ([50c2eab](https://github.com/joshuar/go-hass-agent/commit/50c2eab4951eef19c5591eba1832362c5faaad24))
* **linux/hwmon:** :lipstick: better hwmon sensor naming ([ae5aa9e](https://github.com/joshuar/go-hass-agent/commit/ae5aa9e7d0a0a2e45638237d209e55b04853da6d))
* **linux/hwmon:** :sparkles: expose the sysfs path for the hwmon sensor ([4a198fa](https://github.com/joshuar/go-hass-agent/commit/4a198fafb24ec0f2c25a9dff7a14f6dd8b94f024))


### Bug Fixes

* **linux:** :bug: disk IO rate sensors should be marked as diagnostic sensors ([8ae0ffe](https://github.com/joshuar/go-hass-agent/commit/8ae0ffee348b9378e4ae34a8e871639b69b641c7))
* **linux:** :bug: send lid state sensor on startup ([b1e2aea](https://github.com/joshuar/go-hass-agent/commit/b1e2aeaa8c9b9a2463a22771d8783182dfbaeb4c))
* **linux:** :pencil2: fix warning message if desktop sensors are unavailable ([1005e35](https://github.com/joshuar/go-hass-agent/commit/1005e357c3b076b272019ed350217def7cdad1fa))
* **linux:** :zap: improve lock handling for running apps sensor ([ed50a68](https://github.com/joshuar/go-hass-agent/commit/ed50a685ef5e7f54a939f498240a04a29c92230c))
* **linux:** :zap: rework network sensor code to attempt to avoid race conditions ([e91f389](https://github.com/joshuar/go-hass-agent/commit/e91f38929f037fbc989e96e1b58e898eedad8b97))
* **linux:** handle hwmon chips with the same name ([16f56dd](https://github.com/joshuar/go-hass-agent/commit/16f56dda42ca9d3d525c2d0d59446549a7a7c5b0))
* **logging:** :bug: try to create the directory for log file storage. change error msgs to warn if cannot ([596f6e4](https://github.com/joshuar/go-hass-agent/commit/596f6e4e6d05684ef883901d9659d6469721c3cb))
* **scripts:** :art: capture and return script parser errors ([949bef5](https://github.com/joshuar/go-hass-agent/commit/949bef5483eb39f9a2627dc4c27ae6ba9865f647))

## [7.3.1](https://github.com/joshuar/go-hass-agent/compare/v7.3.0...v7.3.1) (2024-04-17)


### Bug Fixes

* **agent:** :bug: ensure .desktop file is valid and keep previous change for desktop environment display ([1aa6fef](https://github.com/joshuar/go-hass-agent/commit/1aa6feffe42065e0d9baf45320cfe4e35c7d98e6))
* **agent:** :bug: ensure agent shows up in the menus of more desktop environments ([9e93473](https://github.com/joshuar/go-hass-agent/commit/9e9347399f90b3d3c5c649b55f817568c56fcb5e))
* **agent:** :bug: reconnect to MQTT on disconnect and fix issue with MQTT commands not working ([cfce533](https://github.com/joshuar/go-hass-agent/commit/cfce53368c73fc335ca578b2666245594b449815))
* **hass:** :sparkles: switch registry implementations to fix unregisterable sensors ([c31b06c](https://github.com/joshuar/go-hass-agent/commit/c31b06c574fc149c0498f2cdcd87c3e953085058))
* **linux:** :loud_sound: better warning message when app sensors cannot run ([c640bac](https://github.com/joshuar/go-hass-agent/commit/c640bac0a5502a07bc2b84bcd30c154c2af74c99))
* **linux:** sending offline event when interface goes offline ([8ed1a8a](https://github.com/joshuar/go-hass-agent/commit/8ed1a8a619ece8f280b96324316faf3b8c609a7a))
* **scripts:** :bug: prevent invalid script causing agent crash ([07d3e0e](https://github.com/joshuar/go-hass-agent/commit/07d3e0ea36a4a3836641e4585964a4278bbed62a))


### Performance Improvements

* **preferences:** :zap: replace golang.org/x/sync/errgroup with github.com/sourcegraph/conc/pool ([1c814ae](https://github.com/joshuar/go-hass-agent/commit/1c814aea86a9f5c9af39f39f24f02805b54d1833))

## [7.3.0](https://github.com/joshuar/go-hass-agent/compare/v7.2.0...v7.3.0) (2024-04-09)


### Features

* **agent:** :sparkles: add support for screensaver control under Xfce desktop ([78e560c](https://github.com/joshuar/go-hass-agent/commit/78e560c86244189fd5f82a58b830a8702d327893))
* **agent:** :sparkles: support stateless MQTT ([a8e886a](https://github.com/joshuar/go-hass-agent/commit/a8e886a3105c016b4821979651be79a908d1b354))
* **linux:** added laptop lid sensor ([0d92428](https://github.com/joshuar/go-hass-agent/commit/0d92428de0d19166575a0b75c314425dcb9e592f))


### Bug Fixes

* **agent:** :bug: correct D-Bus path for Xfce screensaver control ([9650059](https://github.com/joshuar/go-hass-agent/commit/9650059e1f34e5310e3400cc5dd94a638eeee900))


### Performance Improvements

* **agent:** :zap: improve reliability and error handling of websocket connection ([344e78a](https://github.com/joshuar/go-hass-agent/commit/344e78a732e75c6168fb2fe3f894e5c79392b8e0))
* **hass:** :sparkles: rework request error handling ([1d9d372](https://github.com/joshuar/go-hass-agent/commit/1d9d372df059975cfa4b567c1087e100914f2197))

## [7.2.0](https://github.com/joshuar/go-hass-agent/compare/v7.1.0...v7.2.0) (2024-03-03)


### Features

* **agent:** :sparkles: add suspend and hibernate control via MQTT ([f1678ea](https://github.com/joshuar/go-hass-agent/commit/f1678ea83a03dfe32460e7eddd958d16a7e2d8a4))
* **agent:** :sparkles: allow overriding URL for API requests ([3d1c9d9](https://github.com/joshuar/go-hass-agent/commit/3d1c9d9d1c266d315e5418deeaa5a16928255432))
* **agent:** :sparkles: set the auto-detected server to a default value for convienience ([f39ef5b](https://github.com/joshuar/go-hass-agent/commit/f39ef5b0b1c0d99eb612fff0e473c374e5254fad))


### Bug Fixes

* **dbusx:** :bug: avoid nil pointer access when busRequest exists but bus conn doesn't ([6f69316](https://github.com/joshuar/go-hass-agent/commit/6f69316d0540ce05aa882281bc3bdf63c4111903))
* **hass:** :bug: handle APIError or HTTP Error response more gracefully ([68b18dc](https://github.com/joshuar/go-hass-agent/commit/68b18dc016cf4335a8b50751c757096b1cf0cd11))
* **hass:** :bug: handle uunknown error ([103ebeb](https://github.com/joshuar/go-hass-agent/commit/103ebebf4445fe07ceed46af792395324fa690df))
* **hass:** :bug: support string or int code return for API errors ([c0ebed7](https://github.com/joshuar/go-hass-agent/commit/c0ebed7a7b260f713089e01e3b4d1efc7e971165))
* **hass:** :lock: don't show the URL in trace logging output ([aac3ef8](https://github.com/joshuar/go-hass-agent/commit/aac3ef83245e6bdb8146a58661fd5d09bf8f7da3))
* **hass:** :zap: increase request timeout to a more realistic time to wait for requests to complete ([4a48ab3](https://github.com/joshuar/go-hass-agent/commit/4a48ab31f4e856771bedb31e608bd530db78cad9))

## [7.1.0](https://github.com/joshuar/go-hass-agent/compare/v7.0.1...v7.1.0) (2024-02-26)


### Features

* **agent:** :sparkles: add agent reset command ([853ee60](https://github.com/joshuar/go-hass-agent/commit/853ee60449a41213e0784b23687d0cd1f7ecdb74))
* **agent:** arbitrary dbus commands via MQTT (thanks [@jaynis](https://github.com/jaynis)!) ([7204181](https://github.com/joshuar/go-hass-agent/commit/7204181b746ee9602f83af02ef428cf98ed37a60))

## [7.0.1](https://github.com/joshuar/go-hass-agent/compare/v7.0.0...v7.0.1) (2024-02-20)


### Bug Fixes

* **agent:** :bug: load preferences from file to get MQTT preferences ([6f92a75](https://github.com/joshuar/go-hass-agent/commit/6f92a7572da11d7bf1bde2b6f277268a58f5b3b2))

## [7.0.0](https://github.com/joshuar/go-hass-agent/compare/v6.5.0...v7.0.0) (2024-02-17)


### ⚠ BREAKING CHANGES

* **dbusx:** The dbusx package now uses Go generics for some functions, to combine both fetching or setting a value or property as the required type.
* Major refactor of requests code with internal breaking changes. Migrate from `requests` package to `resty` package. This allows exposing more details about the response from Home Assistant, providing cleaner response handling code. In addition, refactor code to migrate tracker and request code into the hass package, keeping the sensor code as a distinct package for now.
* Legacy agent config package has been removed and replaced with preferences package. This breaks upgrades from all versions besides the last release in the previous major version series. **Users upgrading from older releases should first upgrade to the latest version of the last major release before this release, then upgrade to this release.**

### Features

* :alembic: add ability to run a trace/heap/cpu profile over execution lifetime ([9b73cd8](https://github.com/joshuar/go-hass-agent/commit/9b73cd8094159788f9bd14d95037bbd0a96deab4))
* :recycle: rework sensor registry to abstract from sensor tracker ([a88a04a](https://github.com/joshuar/go-hass-agent/commit/a88a04a9ab73cf918761cb7c42d75dda43e58eea))
* **agent:** :arrow_up: update for latest go-hass-anything ([db884fe](https://github.com/joshuar/go-hass-agent/commit/db884fe25c249e21fb5a19e91a4ed1b3e3dfcc69))
* **dbusx:** use generics to simplify dbusx usage ([45335c4](https://github.com/joshuar/go-hass-agent/commit/45335c4cf5b761bd4fd1789d6fe154cb644f95f6))
* **device:** :sparkles: migrate external ip checker to resty package ([4894fef](https://github.com/joshuar/go-hass-agent/commit/4894fefe72b1e4894832276545ff2629e49ce392))
* **hass:** :sparkles: API response rewrite ([a979728](https://github.com/joshuar/go-hass-agent/commit/a979728ee48ed7cf1bdfb6671fe7b2c7935edf2b))
* **hass:** :sparkles: new functions to retrieve entities from Home Assistant ([a3d0fc6](https://github.com/joshuar/go-hass-agent/commit/a3d0fc66c500f77d9ed2442a237669dfb91b6545))
* **hass:** :sparkles: utilise new ExecuteRequest function ([dff2e83](https://github.com/joshuar/go-hass-agent/commit/dff2e835fc42594a8b5523f5d2f4e5cb2ec2c86d))
* remove config and replace with preferences ([630d4e6](https://github.com/joshuar/go-hass-agent/commit/630d4e61c074c9d3bf10de23b9bc77eaa0715ae5))
* **ui:** :lipstick: show dialogs for success/failure of saving preferences ([a2ab9c2](https://github.com/joshuar/go-hass-agent/commit/a2ab9c25b768246f0b4432d7f7308c9bb5414b51))
* **ui:** :sparkles: show extra details in about window ([e8277cc](https://github.com/joshuar/go-hass-agent/commit/e8277cc4ff2526bfb6155f918e618050a5a55372))


### Bug Fixes

* :sparkles: log file name set in cmd package ([11c4dd1](https://github.com/joshuar/go-hass-agent/commit/11c4dd1a73c52a2438547b6289215350d1b7767e))
* :zap: only retry if the server is overloaded by default ([23f214f](https://github.com/joshuar/go-hass-agent/commit/23f214f4aaa1bea6908598af9db56705f43f80eb))
* **agent,hass:** :bug: fix registration flow ([486890c](https://github.com/joshuar/go-hass-agent/commit/486890cf57724725941b9e98844b5365efc599cd))
* **agent:** :bug: pass appropriate context to runners ([d96a4e5](https://github.com/joshuar/go-hass-agent/commit/d96a4e55bfdb489bc41d1874012e490ec0dd1fa4))
* **agent:** :recycle: clean up context creation in agent ([cad5d56](https://github.com/joshuar/go-hass-agent/commit/cad5d561f7c2f2dde6ccfa149a7d2ce933e6c4f5))
* **device:** :bug: remove spew ([8b81d77](https://github.com/joshuar/go-hass-agent/commit/8b81d77c37d58568933da5924868e4d35caa49a0))
* **hass:** :bug: ensure registry directory is created if it does not exist ([e33a4d4](https://github.com/joshuar/go-hass-agent/commit/e33a4d4bc0b8455e017ebfc18553f47823650da2))
* **hass:** :bug: fix naming of device class values presented to Home Assistant ([ad4a73a](https://github.com/joshuar/go-hass-agent/commit/ad4a73a0648d122cf1a7b97ba05f03771b1e19d8))


### Performance Improvements

* **hass:** :zap: remove unneccesary goroutine usage for ExecuteRequest ([8505455](https://github.com/joshuar/go-hass-agent/commit/8505455ab4f059bce0fa03d3684925516bf45223))


### Code Refactoring

* major requests refactor ([24097f3](https://github.com/joshuar/go-hass-agent/commit/24097f34c040dc5a79b78eba727557917da39419))

## [6.5.0](https://github.com/joshuar/go-hass-agent/compare/v6.4.0...v6.5.0) (2024-02-06)


### Features

* :sparkles: major config rewrite ([680bee1](https://github.com/joshuar/go-hass-agent/commit/680bee1b074c4a65fee4f2312b8003a5129148c4))
* **cmd:** :art: move long command descriptions to embedded text files ([58c2305](https://github.com/joshuar/go-hass-agent/commit/58c2305b776057a5f4ce0e3c7f945d230d746c23))


### Bug Fixes

* :bug: registration flow for new install ([f71d7c6](https://github.com/joshuar/go-hass-agent/commit/f71d7c61f0deb96842b2f6d99b93290fb6c5c5af))
* **agent:** :bug: check for mqtt enabled ([8c2e5f0](https://github.com/joshuar/go-hass-agent/commit/8c2e5f065051ad90fcaa2ba7ac1a825d5f74f991))
* **config:** :bug: handle mqtt config migration quirk ([c4824f3](https://github.com/joshuar/go-hass-agent/commit/c4824f36bc4c3cfc917861c80c408a1598c95ab7))

## [6.4.0](https://github.com/joshuar/go-hass-agent/compare/v6.3.1...v6.4.0) (2024-01-29)


### Features

* **agent:** :art: MQTT agent adjustments ([b094c4a](https://github.com/joshuar/go-hass-agent/commit/b094c4a081db82c628e72097f8bacc4c039ffa50))
* **agent:** :sparkles: control the agent via MQTT ([5756092](https://github.com/joshuar/go-hass-agent/commit/5756092f3b76f3cd2ec10e24d6bdf85c38f767bf))
* **agent:** :sparkles: map mqtt settings to go-hass-agent package settings ([a3dee24](https://github.com/joshuar/go-hass-agent/commit/a3dee246feb333a2e3edcb5e25d3e118bccd110f))
* **cmd,agent:** :sparkles: agent rework ([8ab63e2](https://github.com/joshuar/go-hass-agent/commit/8ab63e2b9ff128f7ed887be0b37f7aff22a35e6d))
* **config:** :sparkles: Export an AppURL config option ([0076019](https://github.com/joshuar/go-hass-agent/commit/0076019e15c67aede3ad6d80f156102deb0ea020))
* **linux:** :sparkles: add CPU Usage % sensor ([6fdb91b](https://github.com/joshuar/go-hass-agent/commit/6fdb91be740542ae2c7c35c89f65b3bf6c417bff))
* **linux:** :sparkles: add memory and swap usage % sensors ([3a7ca08](https://github.com/joshuar/go-hass-agent/commit/3a7ca08ed36d6e2338fe3ab79d6037879a67c4fe))
* **ui,agent,config:** :sparkles: UI overhaul ([ded576b](https://github.com/joshuar/go-hass-agent/commit/ded576b44825e9f5d64494440af3bc99afe58ce0))


### Bug Fixes

* **agent:** :art: device context abstraction ([878438b](https://github.com/joshuar/go-hass-agent/commit/878438b3920bb5cf20567a9a670001546e86cdf4))
* **agent:** :bug: correct registration logic ([569091d](https://github.com/joshuar/go-hass-agent/commit/569091d2c74f8388439b851e260c85f6bd978e0b))
* **agent:** :bug: fix race condition where agent exits before workers start ([1976238](https://github.com/joshuar/go-hass-agent/commit/1976238d1835fcf336bf991d567ae5c7cf910ced))
* **linux/dbusx:** :bug: check nil struct not attribute ([ed4d3be](https://github.com/joshuar/go-hass-agent/commit/ed4d3bef0275f78d96376d96e083c7260c82ac2c))

## [6.3.1](https://github.com/joshuar/go-hass-agent/compare/v6.3.0...v6.3.1) (2024-01-22)


### Bug Fixes

* **linux:** :sparkles: ensure sensors have appropriate icon, device class and state class ([1756f2c](https://github.com/joshuar/go-hass-agent/commit/1756f2c2601578f5a9524e5c68aab392dcd231d6))
* **linux:** :sparkles: support new sensor types exposed via hwmon ([eadacba](https://github.com/joshuar/go-hass-agent/commit/eadacba7ac4e8c69153a0292732e844b267a36e0))
* **linux/hwmon:** :sparkles: expose alarm and intrusion as separate sensors ([5e19a48](https://github.com/joshuar/go-hass-agent/commit/5e19a48ed7312c644446291f5fbb52edc667cbf4))

## [6.3.0](https://github.com/joshuar/go-hass-agent/compare/v6.2.0...v6.3.0) (2024-01-21)


### Features

* **linux:** :sparkles: switch to using hwmon package to get hardware sensors ([a8360c2](https://github.com/joshuar/go-hass-agent/commit/a8360c2c7408e8272725f52580059a49c11bb4ac))
* **linux/hwmon:** :sparkles: add a hwmon package to retrieve all hardware sensors from the kernel ([cf50826](https://github.com/joshuar/go-hass-agent/commit/cf508266438a70c8545ed64e5ae921484cdde993))
* **linux/hwmon:** :sparkles: add an example usage ([174840f](https://github.com/joshuar/go-hass-agent/commit/174840f1334305d9b858df9834e2bed7717f212a))
* **linux/hwmon:** :sparkles: add units output to sensors ([8663945](https://github.com/joshuar/go-hass-agent/commit/8663945f4c69c7573d0ab6ae42d9916b843ead46))
* **linux/hwmon:** :sparkles: apply appropriate scale to sensor values ([bdf3e82](https://github.com/joshuar/go-hass-agent/commit/bdf3e8252a887a1b900f88c45b499b2735e42f16))
* **linux/hwmon:** :sparkles: expose sensor type ([ad970be](https://github.com/joshuar/go-hass-agent/commit/ad970be87753013af7610646ae740bb012584e33))
* **vscode:** :sparkles: add additional conventional commit scopes ([b7d439c](https://github.com/joshuar/go-hass-agent/commit/b7d439c2eb86d935fa1343c98937bf6e4e8688c9))


### Bug Fixes

* **linux/hwmon:** :bug: remove race condition when fetching sensors ([1262c3f](https://github.com/joshuar/go-hass-agent/commit/1262c3f46a8907f04aa83fa0ab0aa02b2d769707))
* **linux/hwmon:** :zap: improve hwmon code ([dee17e4](https://github.com/joshuar/go-hass-agent/commit/dee17e43a26e66ff9dde92f4c41a0abba2c213d3))
* **linux/hwmon:** :zap: reduce struct memory usage ([3f452f6](https://github.com/joshuar/go-hass-agent/commit/3f452f65e004b4adec67f8a7ccb9cdeffebeed6b))

## [6.2.0](https://github.com/joshuar/go-hass-agent/compare/v6.1.2...v6.2.0) (2024-01-14)


### Features

* **agent:** :fire: remove unused and unnecessary info command ([dcc0a31](https://github.com/joshuar/go-hass-agent/commit/dcc0a316e4b62b881da1eef3f30810f196ca57e8))
* **agent:** :sparkles: add error types for use by config code ([961f98f](https://github.com/joshuar/go-hass-agent/commit/961f98f73f3591691d6ceb1d36ecd7dd49cb8a94))
* **agent:** :sparkles: Allow embedding config interface in context ([058950d](https://github.com/joshuar/go-hass-agent/commit/058950d91e580d5eb2ece9d57589bcab87650497))
* **agent:** :zap: simplify config upgrade and validation process ([7fcf14b](https://github.com/joshuar/go-hass-agent/commit/7fcf14bdc8087d8760ce88f9137eea4c72e78457))


### Bug Fixes

* **agent:** :building_construction: wrap workers, scripts and notifications in goroutine with waitgroup ([e593dff](https://github.com/joshuar/go-hass-agent/commit/e593dff2f936eeccff721eff9818dff5321244df))
* **container:** ensure agent runs as a non-privleged user inside a container ([1a3168f](https://github.com/joshuar/go-hass-agent/commit/1a3168f45fc218ebabd5a68dd9f37171bc10cd86))

## [6.1.2](https://github.com/joshuar/go-hass-agent/compare/v6.1.1...v6.1.2) (2024-01-04)


### Bug Fixes

* **agent:** improve warning messages about windowing/UI environment ([222c3ab](https://github.com/joshuar/go-hass-agent/commit/222c3abe2b672ddc29c8d49910d2e4f280ad8e5a))
* **linux:** protect against potential map read/write race condition ([26be3af](https://github.com/joshuar/go-hass-agent/commit/26be3afd63831d618cd2c28507940afeade05353))

## [6.1.1](https://github.com/joshuar/go-hass-agent/compare/v6.1.0...v6.1.1) (2023-12-27)


### Bug Fixes

* **linux:** capture more possible screen lock events ([775a9ab](https://github.com/joshuar/go-hass-agent/commit/775a9ab64ce0c01747a930311a2231b01e10d325))

## [6.1.0](https://github.com/joshuar/go-hass-agent/compare/v6.0.3...v6.1.0) (2023-12-20)


### Features

* **linux:** monitor for battery devices being added/removed ([bde0b2e](https://github.com/joshuar/go-hass-agent/commit/bde0b2ea43d4be530665cef3837b0201daa94bb4))


### Bug Fixes

* **linux:** adjust log levels for some messages ([514bf31](https://github.com/joshuar/go-hass-agent/commit/514bf3121de74fdb371f04b7f7712661e17cb717))
* **linux:** better detection of screenlock D-Bus signal ([fe10af0](https://github.com/joshuar/go-hass-agent/commit/fe10af033c83560b2c15f1c1d5bf58046ed9ab7c))
* **linux:** better naming of battery sensors ([e5f67a1](https://github.com/joshuar/go-hass-agent/commit/e5f67a14521a95a5308d6c929d9631c34dc59d4e))
* **linux:** ensure initially added battery devices are tracked correctly ([20c2574](https://github.com/joshuar/go-hass-agent/commit/20c2574dd9b1a80b51f0e189cbe57369c0b78085))
* **tracker/registry:** better type safety ([ce1afe1](https://github.com/joshuar/go-hass-agent/commit/ce1afe10e6adbf52a288daa7c39383773a94c08c))

## [6.0.3](https://github.com/joshuar/go-hass-agent/compare/v6.0.2...v6.0.3) (2023-12-17)


### Bug Fixes

* **dbushelpers:** adjust logging levels for soft errors ([1966408](https://github.com/joshuar/go-hass-agent/commit/1966408be6ab5226ea89e92199270eae7f9b2601))
* **linux:** correct batteryLevels and batteryStates values for batterySensor ([2c53790](https://github.com/joshuar/go-hass-agent/commit/2c53790465f641c7efb44f1fa3f4f334630aa212))
* **tracker:** more flexible channel return ([5b890f1](https://github.com/joshuar/go-hass-agent/commit/5b890f167b472bd8280b8d90079fa4c1e2f28e1b))

## [6.0.2](https://github.com/joshuar/go-hass-agent/compare/v6.0.1...v6.0.2) (2023-12-16)


### Bug Fixes

* **agent:** remove unused app settings for MQTT for now ([274d4dc](https://github.com/joshuar/go-hass-agent/commit/274d4dcf69d8715194ec6333454ca4df571c8737))
* **linux:** rework batterySensor code to reduce complexity and improve safety ([a1349a6](https://github.com/joshuar/go-hass-agent/commit/a1349a6db07bbc15f1b7b623ac254290ae594273))

## [6.0.1](https://github.com/joshuar/go-hass-agent/compare/v6.0.0...v6.0.1) (2023-12-13)


### Bug Fixes

* **linux:** screen lock sensor type casting issue ([a975f04](https://github.com/joshuar/go-hass-agent/commit/a975f04edc23f483313caf7dfac34b47a2b342d1))

## [6.0.0](https://github.com/joshuar/go-hass-agent/compare/v5.3.1...v6.0.0) (2023-12-13)


### ⚠ BREAKING CHANGES

* **agent:** drop upgrade support for versions < 3.0.0
* **dbushelpers:** improve code safety and logic
* **agent,linux:** return a channel for sensor updates from updater funcs

### Features

* **agent:** drop upgrade support for versions &lt; 3.0.0 ([a33167e](https://github.com/joshuar/go-hass-agent/commit/a33167e297ebcf253520f42692a92695816e246d))
* **dbushelpers:** improve code safety and logic ([c033587](https://github.com/joshuar/go-hass-agent/commit/c033587bdf66783fb01c3452eb9980259518d146))
* **hass/api:** rework sensor response parsing to simplify code ([bc935e5](https://github.com/joshuar/go-hass-agent/commit/bc935e5d649d8f0fb1a584aeb721204ba36dde97))


### Bug Fixes

* **linux,agent:** spelling of PowerProfileUpdater function ([f5a63b8](https://github.com/joshuar/go-hass-agent/commit/f5a63b88bdb100d12a0e748da5174cc8dce8edaa))
* **linux:** batterySensor should inherit linuxSensor ([f1d09ca](https://github.com/joshuar/go-hass-agent/commit/f1d09ca9e3a4361238b445e0c2c9837905dcb1bf))
* **linux:** ensure type assertion is checked ([dc312f5](https://github.com/joshuar/go-hass-agent/commit/dc312f52259c9732fe2f34f9f378fee569b1b9c2))
* **linux:** more type assertion checks ([46ab539](https://github.com/joshuar/go-hass-agent/commit/46ab5396e763d43f99aee30a6ba403f8ae625a2d))
* **linux:** portal detection ([e3b2606](https://github.com/joshuar/go-hass-agent/commit/e3b2606cdad6fa8f94712f8b9f19da8a5197f35f))
* **linux:** protect against divide by zero in networkStatsSensor ([b79d0b2](https://github.com/joshuar/go-hass-agent/commit/b79d0b2f285fcfec4a03176b19400f49fd4e0d90))
* **linux:** protect against error in type assertion for batterySensor ([8eff883](https://github.com/joshuar/go-hass-agent/commit/8eff8830ef4ce8e2c53145087c19821f63a96d6d))
* **linux:** protect type assertion for username list generation ([3a3e4af](https://github.com/joshuar/go-hass-agent/commit/3a3e4af69a236d4dd522bf04f93bebba8e7e7e2d))
* **linux:** protect type assertions for wifi sensors ([c3e85ac](https://github.com/joshuar/go-hass-agent/commit/c3e85acd29289c01d0c23c7d2a14dd3bd94d79c4))
* **linux:** type assertion check for generating power state icon ([f1f0099](https://github.com/joshuar/go-hass-agent/commit/f1f0099ea9dada78135051a519dcab1b0c7d0851))
* **linux:** type assertion check for generating screen lock icon ([50cd304](https://github.com/joshuar/go-hass-agent/commit/50cd3042c0b6dc50dd49c57d36e3d4ec585c68d0))
* **linux:** variable shadows import ([860b6a8](https://github.com/joshuar/go-hass-agent/commit/860b6a8a12d870da1cf773d69ffcc1832e3dea69))


### Code Refactoring

* **agent,linux:** return a channel for sensor updates from updater funcs ([c64c369](https://github.com/joshuar/go-hass-agent/commit/c64c36959b05925549520d533faaf9731b3dbb96))

## [5.3.1](https://github.com/joshuar/go-hass-agent/compare/v5.3.0...v5.3.1) (2023-12-06)


### Bug Fixes

* **linux:** protect against nil value panic in batterySensor ([94af4da](https://github.com/joshuar/go-hass-agent/commit/94af4daca7ebe4c71a2dffde05480b8618ccc8ad))

## [5.3.0](https://github.com/joshuar/go-hass-agent/compare/v5.2.0...v5.3.0) (2023-12-02)


### Features

* **dbushelpers:** simpler signal removal logic ([b185320](https://github.com/joshuar/go-hass-agent/commit/b1853203e366f322339bee42e5d01d8a60082069))


### Bug Fixes

* **agent:** better logging around finding scripts ([bcf41f9](https://github.com/joshuar/go-hass-agent/commit/bcf41f919ef34c12fdbfa785dec2e12bb6b7b7b9))
* **dbushelpers:** remove not useful debug log messages ([a78118c](https://github.com/joshuar/go-hass-agent/commit/a78118cb20b2751fc3481e3b7e69f8faea70a751))
* **linux:** adjust power state D-Bus watch ([d1eefeb](https://github.com/joshuar/go-hass-agent/commit/d1eefeb9e65882ffa137be0c0d89e537a8836f4c))
* **linux:** clean up active app D-Bus watch ([6380b6b](https://github.com/joshuar/go-hass-agent/commit/6380b6bfc81a1b2c3f9b8174701bfe61639368cf))
* **linux:** clean up location sensor D-Bus watch ([6fad74e](https://github.com/joshuar/go-hass-agent/commit/6fad74eb8469753b2cc4bce5149ce1894ce007ef))
* **linux:** clean up users D-Bus watch ([fd14a16](https://github.com/joshuar/go-hass-agent/commit/fd14a1675469662ea35a49adf7162561c1e0c23c))
* **linux:** improved network connection sensor detection and error handling ([f456b55](https://github.com/joshuar/go-hass-agent/commit/f456b55dc2cd425a40489125e79c7a15e15b6e0e))
* **linux:** power profile sensor reporting incorrect state ([5591998](https://github.com/joshuar/go-hass-agent/commit/55919983626dd1fa18e36f3384689b9eff5e59dc))
* **linux:** screen lock sensor improved logic and error checking ([dfa1123](https://github.com/joshuar/go-hass-agent/commit/dfa112390d63e095bc6efdd741d3848be33ba860))

## [5.2.0](https://github.com/joshuar/go-hass-agent/compare/v5.1.2...v5.2.0) (2023-11-27)


### Features

* **agent:** add script sensors ([ece4ddd](https://github.com/joshuar/go-hass-agent/commit/ece4ddda986179000cf4f423286bafb2a732f9d9))
* **cmd:** auto-detect whether to run in headless mode or not ([7c77032](https://github.com/joshuar/go-hass-agent/commit/7c77032ee9eb4b3b705944ed4c921b4f44cb213d))
* **scripts:** support TOML output ([2a67c32](https://github.com/joshuar/go-hass-agent/commit/2a67c32d2dbdfd38ede23679220aa19681a3dad4))

## [5.1.2](https://github.com/joshuar/go-hass-agent/compare/v5.1.1...v5.1.2) (2023-11-19)


### Miscellaneous Chores

* release 5.1.2 ([abf0a85](https://github.com/joshuar/go-hass-agent/commit/abf0a850bf7f200a63d249e14a6acb697d575dfd))

## [5.1.1](https://github.com/joshuar/go-hass-agent/compare/v5.1.0...v5.1.1) (2023-11-18)


### Bug Fixes

* **hass:** better handling of potential nil values ([71667e7](https://github.com/joshuar/go-hass-agent/commit/71667e750da753ccb0ba610267fdf78eb6c402ed))
* **linux:** alternative approach to tracking screen lock state ([be67e53](https://github.com/joshuar/go-hass-agent/commit/be67e5381912868d0fe9a91d45fdf374bcbdaf33))
* **linux:** ensure power state is sent immediately ([70d60d5](https://github.com/joshuar/go-hass-agent/commit/70d60d5d758383ad1a56a64824a3ac4aeff1eb10))
* **linux:** remove call trace on log message ([a79058b](https://github.com/joshuar/go-hass-agent/commit/a79058bbf2045b5957e72a0028f5d82bc2db5e7f))
* **linux:** return a nil dbus.Variant if prop not retrieved ([6db5c79](https://github.com/joshuar/go-hass-agent/commit/6db5c79a95c2e6dce26407d3f2cd93e37a5c05a5))
* **linux:** simplify watch for screen lock sensor ([d7c4399](https://github.com/joshuar/go-hass-agent/commit/d7c43999b7955f4cccb25964dd95c4f7c359ba45))

## [5.1.0](https://github.com/joshuar/go-hass-agent/compare/v5.0.2...v5.1.0) (2023-11-08)


### Features

* **build:** better container support ([ada30ec](https://github.com/joshuar/go-hass-agent/commit/ada30ec115f2ed4531beb78159b6f5a6199649a3))
* **linux:** add sensors for tracking shutdown/suspend state via D-Bus ([cbcb8b5](https://github.com/joshuar/go-hass-agent/commit/cbcb8b516743013f2596bd93c1c3ccd6f6692eae))
* **linux:** change power management sensor to power state sensor ([820a23f](https://github.com/joshuar/go-hass-agent/commit/820a23f8fa9d09657fefa71d5b2fd6a9d7e4b583))
* **linux:** network transfer rate sensors ([851b517](https://github.com/joshuar/go-hass-agent/commit/851b517eb20f27929f1eec42cedb750276de3317))
* **linux:** rework network connections sensor code ([a6f8dfb](https://github.com/joshuar/go-hass-agent/commit/a6f8dfbb51925d48d99a384785352cc780ae361a))
* **tracker:** log a warning if an unknown sensor has been sent ([09604b2](https://github.com/joshuar/go-hass-agent/commit/09604b2997de1850f8c873c77113ed34e9a6e907))


### Bug Fixes

* **linux:** add datasource to network transfer rates sensors ([043c8a4](https://github.com/joshuar/go-hass-agent/commit/043c8a437c7d6a1021a42137a282d90e9e1fc3c8))
* **linux:** better check for no address for connection ([6c66a67](https://github.com/joshuar/go-hass-agent/commit/6c66a67b53aea674823902d60fd7cd0facacdf9b))
* **linux:** clean up D-Bus connection and signals on shutdown ([c815184](https://github.com/joshuar/go-hass-agent/commit/c815184b1cfd6b557936107563344a5a586800a4))
* **linux:** correct tracking of user sessions created/removed ([7f7e01b](https://github.com/joshuar/go-hass-agent/commit/7f7e01bce7d17616f5ce10f24e6775f9b1cef060))
* **linux:** fix name of power state sensor in warning message ([136da9c](https://github.com/joshuar/go-hass-agent/commit/136da9cb0ddd383dad35626e69718f5c2189aa2d))
* **linux:** follow android app and treat wifi sensors as diagnostics ([f06cbc4](https://github.com/joshuar/go-hass-agent/commit/f06cbc43ff86ee224985a058c80bdaf2369fb10f))
* **linux:** handle no address for network connection sensor ([254ad19](https://github.com/joshuar/go-hass-agent/commit/254ad19a0a0dad43a351b27c34dbfaceceb604cd))
* **linux:** make AddWatch non-blocking, add more logging ([053a86b](https://github.com/joshuar/go-hass-agent/commit/053a86b10f5a8e7930d6639697039b7f6e554db6))
* **linux:** network connection sensor state should be diagnostic ([3b42944](https://github.com/joshuar/go-hass-agent/commit/3b429440e2241cee76903396c8726113cdd9cd39))
* **linux:** set values of power management sensors on startup ([7ac5304](https://github.com/joshuar/go-hass-agent/commit/7ac53045703e4681be74f6df44d5300a9053b7b6))

## [5.0.2](https://github.com/joshuar/go-hass-agent/compare/v5.0.1...v5.0.2) (2023-10-28)


### Bug Fixes

* **agent:** should wait for waitgroups ([8c711a0](https://github.com/joshuar/go-hass-agent/commit/8c711a0bd87dda4ce0d42e02bd5776cccfe55bbc))

## [5.0.1](https://github.com/joshuar/go-hass-agent/compare/v5.0.0...v5.0.1) (2023-10-24)


### Bug Fixes

* **agent,hass,device:** better clean-up on agent quit/cancellation ([ec7a7e0](https://github.com/joshuar/go-hass-agent/commit/ec7a7e08e892c866a917cc11ba4bec984cd47e27))

## [5.0.0](https://github.com/joshuar/go-hass-agent/compare/v4.1.1...v5.0.0) (2023-10-12)


### ⚠ BREAKING CHANGES

* **agent:** switch to config powered by Viper

### Features

* **agent:** improve config upgrade and validation ([31561e0](https://github.com/joshuar/go-hass-agent/commit/31561e02dbae01b70e09f8b6b4065c7a1a5b0dfd))
* **agent:** migrate registry as part of Fyne to Viper config migration ([e347c94](https://github.com/joshuar/go-hass-agent/commit/e347c948de18b42bd5382f5c840b8e1315e9cb7c))
* **agent:** switch to config powered by Viper ([cd27058](https://github.com/joshuar/go-hass-agent/commit/cd2705870ce2149d81dc7345c6356d6642e15e18))
* **build:** add a Dockerfile ([1e0bc96](https://github.com/joshuar/go-hass-agent/commit/1e0bc969db009642ec765d51e8e31d8da39cd3dd))
* **linux:** add temperature sensors ([4815197](https://github.com/joshuar/go-hass-agent/commit/4815197df3af66dde03ca466dd6c9d1132489597))
* **linux:** simplify getting hostname and hardware details ([ddff4e2](https://github.com/joshuar/go-hass-agent/commit/ddff4e27f84e16d0870f0236018bd0afe8127117))


### Bug Fixes

* **agent,hass:** fix logic around retrying websocket connection ([6603e06](https://github.com/joshuar/go-hass-agent/commit/6603e061f6887851cd0a8439093bb25a20b8abe4))
* **agent/config:** don't try to migrate already migrated registry ([035430a](https://github.com/joshuar/go-hass-agent/commit/035430a6d0f6f2f6c1aaf6224fb6be513e84efa1))
* **agent/register:** avoid unnecessary config validation ([9ce8a4b](https://github.com/joshuar/go-hass-agent/commit/9ce8a4be8fc14e8c185c8e98d13bd247d9393d27))
* **agent/ui:** remove broken HA version display ([4705a35](https://github.com/joshuar/go-hass-agent/commit/4705a35cc732cf985adcb2229ef2f873ee4d0f6c))
* **agent:** broken registration validation flow ([e3ef8f2](https://github.com/joshuar/go-hass-agent/commit/e3ef8f20bb9181d660b996079e078fe925b47183))
* **agent:** command-line registration flow ([d63752e](https://github.com/joshuar/go-hass-agent/commit/d63752e9f567ceab3338658e47cb558777aa0de6))
* **agent:** continue if config upgrade fails because config does not exist ([c116a8d](https://github.com/joshuar/go-hass-agent/commit/c116a8d5108ae50e6c43e86e8a3cfbb67ee7eadc))
* **agent:** don't use context for linux device creation ([a0d8bb1](https://github.com/joshuar/go-hass-agent/commit/a0d8bb145690556a7c00d68e8c19ae24e4b9952b))
* **agent:** remove the need to import custom viper package in agent package ([833e78c](https://github.com/joshuar/go-hass-agent/commit/833e78c65104ccb54841256df9bcecd763019bdf))
* **agent:** use command-line debugid for config path if specified ([3f8b688](https://github.com/joshuar/go-hass-agent/commit/3f8b6880916e204cfede939c54ec95d2fc5017a1))
* **cmd:** wrap long description onto multiple lines ([f2de9fd](https://github.com/joshuar/go-hass-agent/commit/f2de9fd14617ec0aa42ad42d24b9fb620a651360))
* **linux:** add "temp_" prefix to temp sensor entity ids ([c6480f3](https://github.com/joshuar/go-hass-agent/commit/c6480f3195fcd8e06b83acabd798985649b29ca6))
* **linux:** better handling of unavailable sensors ([f070be7](https://github.com/joshuar/go-hass-agent/commit/f070be7910aab099e90de4a4e5ee714cd6363502))
* **linux:** clean up network connections sensor ([78dc843](https://github.com/joshuar/go-hass-agent/commit/78dc8434596c4d8dc8447cb090ab29bcec42f4dc))
* **linux:** show warning if app sensors could not be enabled ([737f4ed](https://github.com/joshuar/go-hass-agent/commit/737f4ed38e339d5c81c21ab75bfb26544a0499dd))
* **linux:** use string constant for procfs source attribute ([9521449](https://github.com/joshuar/go-hass-agent/commit/952144968b015067d8b1b62d239ddf29e73a92ca))

## [4.1.1](https://github.com/joshuar/go-hass-agent/compare/v4.1.0...v4.1.1) (2023-10-03)


### Bug Fixes

* **linux:** sensor type strings ([29b995c](https://github.com/joshuar/go-hass-agent/commit/29b995c88e0d5ec02ef5eef9e6ebf27e66bdb1ad))


### Miscellaneous Chores

* release 4.1.1 ([6d84656](https://github.com/joshuar/go-hass-agent/commit/6d84656bc24b6d8fdd065ddbec32b3cb80ba65bb))

## [4.1.0](https://github.com/joshuar/go-hass-agent/compare/v4.0.3...v4.1.0) (2023-10-02)


### Features

* **linux:** rewrite D-Bus logic ([fa0e5bc](https://github.com/joshuar/go-hass-agent/commit/fa0e5bc61ace68216b527307f0fdbfaa4c690599))
* **linux:** simplify D-Bus signal matching ([44a0d74](https://github.com/joshuar/go-hass-agent/commit/44a0d742015473709cd2d58ad0e8d491f5f2cf91))


### Bug Fixes

* **agent/ui:** don't do any init of Fyne if running headless ([533a1f2](https://github.com/joshuar/go-hass-agent/commit/533a1f2a24c7c0c6daae05e5ddcacad994b98e11))
* **agent/ui:** remove unused setting ([90d7ce9](https://github.com/joshuar/go-hass-agent/commit/90d7ce93821024885a6d7eeb47b7694d5c7fc136))
* **linux:** fix spelling mistake ([3deea46](https://github.com/joshuar/go-hass-agent/commit/3deea466cb2dca64c43788b05974d54584a16747))

## [4.0.3](https://github.com/joshuar/go-hass-agent/compare/v4.0.2...v4.0.3) (2023-09-27)


### Bug Fixes

* **agent:** tray icon not shown ([d2bcc00](https://github.com/joshuar/go-hass-agent/commit/d2bcc00888a1a0cbca6199fd358be26e802936d4))

## [4.0.2](https://github.com/joshuar/go-hass-agent/compare/v4.0.1...v4.0.2) (2023-09-27)


### Miscellaneous Chores

* release 4.0.2 ([8f116ef](https://github.com/joshuar/go-hass-agent/commit/8f116efc3ba26ca03fad2cbdcc62fca2850f6bc7))

## [4.0.1](https://github.com/joshuar/go-hass-agent/compare/v4.0.0...v4.0.1) (2023-09-27)


### Miscellaneous Chores

* release 4.0.1 ([af2a5a7](https://github.com/joshuar/go-hass-agent/commit/af2a5a79af42836f2d2c59dab14a4a12d162b92a))

## [4.0.0](https://github.com/joshuar/go-hass-agent/compare/v3.3.0...v4.0.0) (2023-09-26)


### ⚠ BREAKING CHANGES

* **agent,hass,tracker:** split UI into own package and more interface usage

### Features

* **agent,device:** change to a variadic list of sensor workers to start ([c6ddac2](https://github.com/joshuar/go-hass-agent/commit/c6ddac25014ef1b83f54243cb3cf0412aa030507))
* **agent,tracker:** move device worker init from tracker to agent ([05b3b1b](https://github.com/joshuar/go-hass-agent/commit/05b3b1b4edfde2cd19312bb981d112c72854b977))
* **agent/config,agent/ui:** add more mqtt prefs. add secret config entry ([28f1ddc](https://github.com/joshuar/go-hass-agent/commit/28f1ddcaa02012597d32aa4ae4dd136678d06bd0))
* **agent/ui:** add a configCheck function for bool config items ([134c876](https://github.com/joshuar/go-hass-agent/commit/134c876955a0156ea8fcea3aeb7ad9c1dd90aa3e))
* **agent/ui:** new validator and placeholder functionality ([98d0cf2](https://github.com/joshuar/go-hass-agent/commit/98d0cf2b1e86cb9a23d9b305f8e3401208c28ea2))
* **agent:** start exposing optional settings for the agent ([dea2cd9](https://github.com/joshuar/go-hass-agent/commit/dea2cd906a63e02bf43034e3b0db27b405c0d37e))
* **cmd:** clean up logging ([316b357](https://github.com/joshuar/go-hass-agent/commit/316b357cabee8a6a14e48d5216736c609cd32484))
* **tracker,device,linux:** move to utilising an interface for updating sensor networkStatsSensor ([af4f0aa](https://github.com/joshuar/go-hass-agent/commit/af4f0aac8ac0a1acbdef194010f25b8c59d37548))


### Bug Fixes

* **agent,hass:** remove Fyne-isms from notification code ([6fee81f](https://github.com/joshuar/go-hass-agent/commit/6fee81f7724a05cc672014960e186739bc2437df))
* **agent:** (hopefully) get some memory savings in sensors table display ([8efbd0b](https://github.com/joshuar/go-hass-agent/commit/8efbd0b460fa068e92648c651ed757ac2cac83ff))
* **agent/ui:** embed tray icon png directly rather than use converted []byte array ([eee1ab1](https://github.com/joshuar/go-hass-agent/commit/eee1ab18b6a46b1f263e3f6de82b40f4f4021622))
* **agent/ui:** only instatiate translator once for UI ([1863c24](https://github.com/joshuar/go-hass-agent/commit/1863c240ed0b453341de24742fb3ffa38c16ed0c))
* **agent/ui:** uncomment code that should be used ([a2131d6](https://github.com/joshuar/go-hass-agent/commit/a2131d676833385b09c28821cbd0a71392a6ff24))
* **agent:** agent struct doesn't need to export any fields ([89fd9e2](https://github.com/joshuar/go-hass-agent/commit/89fd9e29517bef3b2b4b322d511262b3c4edd2e5))
* **agent:** broken registration flow after recent changes ([c227220](https://github.com/joshuar/go-hass-agent/commit/c2272209bf61c84b55b4a3415fab3c15b51eb59c))
* **agent:** don't export version global var ([56fa638](https://github.com/joshuar/go-hass-agent/commit/56fa6386b676186ebef6c695f40b692ccc205f2e))
* **agent:** remove workaround for https://github.com/fyne-io/fyne/issues/3170 ([537e121](https://github.com/joshuar/go-hass-agent/commit/537e1216f8b829ab57410cec245f7dd53390c180))
* **agent:** rework registration/preferences to properly set agent config ([9fd0002](https://github.com/joshuar/go-hass-agent/commit/9fd0002ef7676e8657f9aedfb245e08ed9796f1c))
* **cmd:** debugID argument was ignored after recent logging changes ([ffc40d6](https://github.com/joshuar/go-hass-agent/commit/ffc40d60e647199ededb9a6ad5c1c5276f346c68))
* **device:** don't log transport error when fetching external ip ([c6efde9](https://github.com/joshuar/go-hass-agent/commit/c6efde906622cafb63a7cc3b049154ceec64dc72))
* **device:** signal waitgroup finish properly ([ba309fa](https://github.com/joshuar/go-hass-agent/commit/ba309fa8a32898a3007117d83349f4b966d87f17))
* **device:** wrap polling code in waitgroup ([6e203c2](https://github.com/joshuar/go-hass-agent/commit/6e203c2ad1b82d256cb50c69ffefdd7f8d1341fe))
* **tracker:** use correct context ([8736dc4](https://github.com/joshuar/go-hass-agent/commit/8736dc4d7720f36ef320f37da622394a93c99350))


### Code Refactoring

* **agent,hass,tracker:** split UI into own package and more interface usage ([7eb18bb](https://github.com/joshuar/go-hass-agent/commit/7eb18bb065d424332676ed2f4cdc0a97ebf98564))

## [3.3.0](https://github.com/joshuar/go-hass-agent/compare/v3.2.0...v3.3.0) (2023-09-14)


### Features

* **agent,hass,tracker:** move to interface access to agent config ([9d824ca](https://github.com/joshuar/go-hass-agent/commit/9d824caf099d05b2057ac5af3c986b9a73d5ec37))
* **agent:** sorted sensors table window and update on scroll ([c44aa2a](https://github.com/joshuar/go-hass-agent/commit/c44aa2a4a988f6b77018e9f21b03076731cb4fe4))
* **agent:** values update every n seconds on sensors window ([d3428d6](https://github.com/joshuar/go-hass-agent/commit/d3428d62850c9af56b2c78034c25e94a552c3b8d))


### Bug Fixes

* **hass:** websocket connection should gracefully handle home assistant disconnects/restarts ([1f74f83](https://github.com/joshuar/go-hass-agent/commit/1f74f8356fbc74130c88b91df0f29b47ad5de6a7))

## [3.2.0](https://github.com/joshuar/go-hass-agent/compare/v3.1.0...v3.2.0) (2023-09-10)


### Features

* **linux:** simplify fetching device details for registration ([e2ddf5f](https://github.com/joshuar/go-hass-agent/commit/e2ddf5f22bdf5b2c0b5fb68921b2ebf5b4dc5672))
* **ui:** add headers to sensors window table ([fbcee6e](https://github.com/joshuar/go-hass-agent/commit/fbcee6ee45cc7c32447b4bdef93fd7de1e798d02))

## [3.1.0](https://github.com/joshuar/go-hass-agent/compare/v3.0.3...v3.1.0) (2023-08-27)


### Features

* **linux:** add a sensor to track logged in users and their usernames ([50f76d4](https://github.com/joshuar/go-hass-agent/commit/50f76d4d0394153ca4172e88d33b6a1aea52b1ca))
* **linux:** add new sensors for kernel and distribution version, name ([e9b86a1](https://github.com/joshuar/go-hass-agent/commit/e9b86a1d1ecbef7df5804c1a72b047cb0d640411))


### Bug Fixes

* **linux:** add data source to kernel, distribution version/name sensors ([8243f67](https://github.com/joshuar/go-hass-agent/commit/8243f670a30284de51e3a969f7641f953cb39afc))
* **linux:** add space in name of battery sensors ([fb7b498](https://github.com/joshuar/go-hass-agent/commit/fb7b49810758676dba8b3634fd38944c4058d9fa))

## [3.0.3](https://github.com/joshuar/go-hass-agent/compare/v3.0.2...v3.0.3) (2023-08-22)


### Bug Fixes

* **tracker:** add missing waitgroup.Wait ([0268981](https://github.com/joshuar/go-hass-agent/commit/026898185a8e7a09c38621f18bae61e4c436ffc7))

## [3.0.2](https://github.com/joshuar/go-hass-agent/compare/v3.0.1...v3.0.2) (2023-08-04)


### Bug Fixes

* **linux:** remove unused context ([a25e592](https://github.com/joshuar/go-hass-agent/commit/a25e59253fa5094dd9a950a6a38f9c347bc55259))

## [3.0.1](https://github.com/joshuar/go-hass-agent/compare/v3.0.0...v3.0.1) (2023-07-28)


### Bug Fixes

* **linux:** adjust debug/trace information ([1006240](https://github.com/joshuar/go-hass-agent/commit/1006240200ae6e6c99a100ce9305747cfae51047))
* **linux:** remove duplication of `active_app` and `running_apps` values ([3390b58](https://github.com/joshuar/go-hass-agent/commit/3390b5841b07d04f168ac1fbc9306002038c3e0f))
* **location:** assert type safely ([2f753c4](https://github.com/joshuar/go-hass-agent/commit/2f753c45f082d35876d049d00095a88c81b32783))

## [3.0.0](https://github.com/joshuar/go-hass-agent/compare/v2.0.0...v3.0.0) (2023-07-23)


### ⚠ BREAKING CHANGES

* **tracker:** migration of registry to new format
* **linux:** intialise DBus API internally in linux package
* **agent,tracker:** major tracker rewrite

### Features

* **linux:** add a Data Source attribute to all Linux sensors ([f59d224](https://github.com/joshuar/go-hass-agent/commit/f59d2240d6db3b93c713b998bde1ec02040627bf))
* **tracker:** migration of registry to new format ([c37a793](https://github.com/joshuar/go-hass-agent/commit/c37a79307787a35b36dba102b861982e0f9ee0c6))


### Bug Fixes

* **agent:** bail early if no sensors available when requesting sensors window ([7b8b0d3](https://github.com/joshuar/go-hass-agent/commit/7b8b0d3318a5f390571c22654f55978f441f01c6))
* **agent:** call context cancellation if no need to register ([cf56a8e](https://github.com/joshuar/go-hass-agent/commit/cf56a8e0bfc30d8d9f60f7bbcea352c1d2b48ed2))
* **agent:** headless option crash on startup ([1619ea2](https://github.com/joshuar/go-hass-agent/commit/1619ea26c824e799e2fd6d2a8e62b5eebb7382c4))
* **agent:** if already upgraded, don't try again ([ea9a0b4](https://github.com/joshuar/go-hass-agent/commit/ea9a0b4237f73b73236065c594728eee69484e5a))
* **agent:** make sure to clean up old registry files on upgrade ([8d98899](https://github.com/joshuar/go-hass-agent/commit/8d9889961e8fdcede14480dcb8a318a3b7a9cdcd))
* **agent:** UI fixes ([e96eb0d](https://github.com/joshuar/go-hass-agent/commit/e96eb0d3c21d14b6816faaa0cc64708bdbef9c84))
* **hass:** close of closed channel ([3dd9966](https://github.com/joshuar/go-hass-agent/commit/3dd99665923d6441ddbfb8b419a447e4ae865e98))
* **tracker:** actually use new json registry ([99cee01](https://github.com/joshuar/go-hass-agent/commit/99cee01c238e3015e3f8b3bcef521bddd73dfa21))
* **tracker:** improve handling of disabled state ([5ab5281](https://github.com/joshuar/go-hass-agent/commit/5ab528167efa775299ceb45d3473106bc804f851))


### Code Refactoring

* **agent,tracker:** major tracker rewrite ([a7cb475](https://github.com/joshuar/go-hass-agent/commit/a7cb475b5cd1be0e332b986cd73fd5b7ca9b1b45))
* **linux:** intialise DBus API internally in linux package ([baa6086](https://github.com/joshuar/go-hass-agent/commit/baa608651166c32eb3ab88fe3c22610307c1703c))

## [2.0.0](https://github.com/joshuar/go-hass-agent/compare/v1.4.3...v2.0.0) (2023-07-17)


### ⚠ BREAKING CHANGES

* introduce a Config interface for the api
* **cmd:** remove shorthand flags for debug, profile and terminal
* replace mockery with moq
* major internal rewrite

### Features

* **agent,api,hass,settings:** standardise naming of shared settings ([30ab245](https://github.com/joshuar/go-hass-agent/commit/30ab2455c08815b70c3b8ea0c0aa054e2696468c))
* **agent,api,hass:** new settings package for shared/global settings access ([52601eb](https://github.com/joshuar/go-hass-agent/commit/52601ebfa8ecf800c59c6df5f51ebd74f9d24daf))
* **agent,api:** define an interface for fetching config values for api, ([24369d2](https://github.com/joshuar/go-hass-agent/commit/24369d2a6b7654953f9e93dc772dc93b390a3292))
* **agent,api:** websocket connection fetches needed config from interface ([a972f66](https://github.com/joshuar/go-hass-agent/commit/a972f66b887e867cc5157e46f901d2fcbeecd704))
* **agent,hass,tracker:** simplify HA config interaction ([5ea40b0](https://github.com/joshuar/go-hass-agent/commit/5ea40b0a6713fe34e6ce16687f8f3286c81471d2))
* **agent:** rework registration UI process ([6d907d3](https://github.com/joshuar/go-hass-agent/commit/6d907d32c9df771931be44b768e06e5ce4c29fac))
* **agent:** window resuage ([2de0785](https://github.com/joshuar/go-hass-agent/commit/2de0785150a1aac51f853de42ca996c11542fdba))
* **cmd:** add an option to toggle trace logging ([cff5097](https://github.com/joshuar/go-hass-agent/commit/cff509718e8a318fd7a06a6806c0fc1f49f4e91c))
* introduce a Config interface for the api ([d50b78c](https://github.com/joshuar/go-hass-agent/commit/d50b78cd7da223c3fa90fd1740e8eff6a09efcc0))
* major internal rewrite ([58f9610](https://github.com/joshuar/go-hass-agent/commit/58f961019116802336551d8ccf2ee8c07be349fe))
* **settings,api,hass:** stop sharing settings through a context variable ([291f651](https://github.com/joshuar/go-hass-agent/commit/291f65116d1c3d186e4338862f2a968ae6f67f5a))


### Bug Fixes

* **agent,tracker:** adjust log levels for tracker-related messages ([6938a35](https://github.com/joshuar/go-hass-agent/commit/6938a35d8010dbf8caa93bb29b58871f1d890aae))
* **agent,translations:** adjust levels for misc ui-related messages ([61dc8c6](https://github.com/joshuar/go-hass-agent/commit/61dc8c64288a4005e6fa8a07fe9a5203150c6c59))
* **agent:** abstract app config for more portability ([c580536](https://github.com/joshuar/go-hass-agent/commit/c5805367260136b3d516345b676588ca977f5a1f))
* **agent:** bail on websocket creation if running headless ([e8befd2](https://github.com/joshuar/go-hass-agent/commit/e8befd2cb5af3f0070c59cb3d737a894e7976d5e))
* **agent:** correct logic for retrieving token and server from registration details ([0be94be](https://github.com/joshuar/go-hass-agent/commit/0be94be477e794c72e3dd86f81a3a0a305e57aa2))
* **agent:** expose error messages for config issues, remove call trace ([2ce72c5](https://github.com/joshuar/go-hass-agent/commit/2ce72c5792c5314dd5b1349f8d1440c33b13af89))
* **agent:** rework agent UI ([046c4f4](https://github.com/joshuar/go-hass-agent/commit/046c4f4d1195c24a0c3a17f32fc9db91314c9955))
* **cmd:** remove shorthand flags for debug, profile and terminal ([1fa6add](https://github.com/joshuar/go-hass-agent/commit/1fa6addf335e177a32ddcca28b29388a69db54d2))
* **device:** adjust log levels for external_ip sensor ([e01ab42](https://github.com/joshuar/go-hass-agent/commit/e01ab4228ab2c75de6a1d4cbe0f01c0005421a4b))
* **hass:** adjust log levels for messages in websocket handling ([f4bc216](https://github.com/joshuar/go-hass-agent/commit/f4bc21626b51a32a363c9ed751fcbe33ff88bd52))
* **linux:** adjust log levels for networkConnectionSensor ([5b24fe2](https://github.com/joshuar/go-hass-agent/commit/5b24fe2b30522145780856b1f69c275eb8dd43d4))
* **linux:** adjust logging levels for DBus functions ([27e5654](https://github.com/joshuar/go-hass-agent/commit/27e56547913d4f3dc4fedbdcfe21725b31b9ad51))
* **location:** error log message should be error log level ([da1be00](https://github.com/joshuar/go-hass-agent/commit/da1be00e642f634b647c38a6926fef2c5f859101))
* **tracker:** remove call trace from debug log message for state update ([d3d2976](https://github.com/joshuar/go-hass-agent/commit/d3d29761327b62f0f363cc32a3081b1415f44a0b))


### Tests

* replace mockery with moq ([587d3dc](https://github.com/joshuar/go-hass-agent/commit/587d3dceece7d6c92c654e5f34862cbaafb1d863))

## [1.4.3](https://github.com/joshuar/go-hass-agent/compare/v1.4.2...v1.4.3) (2023-07-07)


### Bug Fixes

* **agent,hass:** store api and websocket urls in config ([2e38d97](https://github.com/joshuar/go-hass-agent/commit/2e38d972c63a3a877cf16c0e2ccf083395c4f391))
* **agent:** add a config upgrade action to remove trailing slash from host ([1be71ad](https://github.com/joshuar/go-hass-agent/commit/1be71adbebc3eab712f52386f416401a4f8fbffd))
* **agent:** additional error checking and code safety ([68f6b90](https://github.com/joshuar/go-hass-agent/commit/68f6b909676115dfa575019fa6d916a4f3813d67))
* **agent:** better logic around closing the app ([b59c1c9](https://github.com/joshuar/go-hass-agent/commit/b59c1c987d6855a29927c9441c430e4be4fe7af9))
* **agent:** ensure agent config satisfies Config interface ([e9aec95](https://github.com/joshuar/go-hass-agent/commit/e9aec950fdb036581a3a4f784b824a35a55babd3))
* **agent:** force registration flow stall ([c1a1511](https://github.com/joshuar/go-hass-agent/commit/c1a1511c984e7a2d20f4853f4095f21d4d412463))
* **agent:** tray menu entries did not display windows and quit did not work ([1bc86b7](https://github.com/joshuar/go-hass-agent/commit/1bc86b7ec55f9a366622b40f1f229311b502797e))
* **config:** adjust interface methods ([bf49018](https://github.com/joshuar/go-hass-agent/commit/bf4901877017665fcf884d0967d0f17802740a29))
* **hass:** check type assertion before using ([8b9e2c3](https://github.com/joshuar/go-hass-agent/commit/8b9e2c3cc874a61b6ff444e529749f6cbbc7cde1))
* **linux:** export LinuxDevice type ([2fd1319](https://github.com/joshuar/go-hass-agent/commit/2fd1319dde2dd029f352338baa55550ff7947b05))
* **sensors:** don't store HA config, just fetch as needed ([ea26f02](https://github.com/joshuar/go-hass-agent/commit/ea26f02d200b50147ac937fcdf17e3a457b170e1))
* **sensors:** handle error in setting registry item ([fd7c15d](https://github.com/joshuar/go-hass-agent/commit/fd7c15d5af476058f8febd9574e0fe10341cb630))
* **sensors:** nil pointer reference condition on new device added to HA ([4d1e6ad](https://github.com/joshuar/go-hass-agent/commit/4d1e6ad2a36b65c4e34233d60b1bea9ac025808f))
* **tracker:** check type assertions in sensor response ([1664606](https://github.com/joshuar/go-hass-agent/commit/16646060aa27418380e66c8f91e7a52c05c7ac79))
* **tracker:** correctly handle setting registration and disabled status ([664be51](https://github.com/joshuar/go-hass-agent/commit/664be512c0ffa0311e7f0668158ef686a79c42ae))
* **tracker:** expose error on db creation ([5f73d8f](https://github.com/joshuar/go-hass-agent/commit/5f73d8f4e08f1192b5b66296b8c97979c42ff285))
* **tracker:** remove spews ([d104593](https://github.com/joshuar/go-hass-agent/commit/d1045936c601a716514b64c5c0dd2b4ef1f88840))

## [1.4.2](https://github.com/joshuar/go-hass-agent/compare/v1.4.1...v1.4.2) (2023-07-03)


### Bug Fixes

* **agent:** remove spew.Dump debugging function ([d1103d8](https://github.com/joshuar/go-hass-agent/commit/d1103d8d59b793b8aa9494d5eb39a54925cfb7a2))
* **agent:** validation of config should accept both old and new formats ([575559a](https://github.com/joshuar/go-hass-agent/commit/575559aa4bf63391ebe3ec5077061d61ae5fe65d))
* **agent:** wrong return type for config option ([dd7b3b7](https://github.com/joshuar/go-hass-agent/commit/dd7b3b76ee42d827bfa85d4d21e3d2d148639067))
* **config,agent,hass:** add an Upgrade function to Config interface ([fb7d405](https://github.com/joshuar/go-hass-agent/commit/fb7d4053fe859fd227e27ac5c56cd3aa79a139ae))

## [1.4.1](https://github.com/joshuar/go-hass-agent/compare/v1.4.0...v1.4.1) (2023-07-03)


### Bug Fixes

* **agent:** wrap tray icon creation in goroutine to avoid block ([5eae020](https://github.com/joshuar/go-hass-agent/commit/5eae020fad0703bd595c1a177c4ed604d2ef4308))
* **cmd:** update info and version commands ([903561f](https://github.com/joshuar/go-hass-agent/commit/903561fdb812e26302cf103140d51776d3acd3b4))
* **hass:** bump websocket dependency and fix breaking changes ([3443df6](https://github.com/joshuar/go-hass-agent/commit/3443df60a5461c18b302ca227c8262f40851cb3c))

## [1.4.0](https://github.com/joshuar/go-hass-agent/compare/v1.3.1...v1.4.0) (2023-07-02)


### Features

* **agent:** rework registration to require a url over hostname/port ([ed3edf6](https://github.com/joshuar/go-hass-agent/commit/ed3edf660fc38976bdf4451afcac3be480f7b7e8))

## [1.3.1](https://github.com/joshuar/go-hass-agent/compare/v1.3.0...v1.3.1) (2023-07-02)


### Bug Fixes

* **agent,hass:** allow hostname or hostname:port for HA server ([b76d112](https://github.com/joshuar/go-hass-agent/commit/b76d112f1c3f94e5b3d6d08da9098ff15654fe8b))

## [1.3.0](https://github.com/joshuar/go-hass-agent/compare/v1.2.6...v1.3.0) (2023-06-29)


### Features

* major re-write for registration process ([6838d29](https://github.com/joshuar/go-hass-agent/commit/6838d29a559a39b28904ff248cd6671752b7215a))


### Bug Fixes

* **agent:** add check to flag existing registration of upgraded agents ([9322500](https://github.com/joshuar/go-hass-agent/commit/9322500836253e6ff641c3865a4b8fb46c27250e))
* **agent:** save `DeviceName` and `DeviceID` with registration ([2356905](https://github.com/joshuar/go-hass-agent/commit/23569054c18d05badf6eae94917fa4fc2f87f402))

## [1.2.6](https://github.com/joshuar/go-hass-agent/compare/v1.2.5...v1.2.6) (2023-06-26)


### Bug Fixes

* **agent:** fix links for creating issue/requesting feature ([01665a6](https://github.com/joshuar/go-hass-agent/commit/01665a6c00c717b331a477c5d350c891b26c3379))

## [1.2.5](https://github.com/joshuar/go-hass-agent/compare/v1.2.4...v1.2.5) (2023-06-22)


### Bug Fixes

* **linux:** potential fix to remove race condition on dbus watches ([83edb5c](https://github.com/joshuar/go-hass-agent/commit/83edb5cb18d1282ca75d04bbbb7835a91c3c0e38))

## [1.2.4](https://github.com/joshuar/go-hass-agent/compare/v1.2.3...v1.2.4) (2023-06-19)


### Bug Fixes

* **sensors:** revert attribute code and clean-up error message ([791a3f2](https://github.com/joshuar/go-hass-agent/commit/791a3f212ac6df948ab2340f36b141c5178ccaba))

## [1.2.3](https://github.com/joshuar/go-hass-agent/compare/v1.2.2...v1.2.3) (2023-06-19)


### Bug Fixes

* **sensors,agent:** refactor sensor tracker to avoid data races ([319a8fc](https://github.com/joshuar/go-hass-agent/commit/319a8fce413819d12a0412830e44a5c114a62840))

## [1.2.2](https://github.com/joshuar/go-hass-agent/compare/v1.2.1...v1.2.2) (2023-06-14)


### Bug Fixes

* **agent:** add validation to token entry for registration ([9af1984](https://github.com/joshuar/go-hass-agent/commit/9af198432953dc302bab797b5ca711c40aaf2ca5))
* **agent:** fix error message ([4fba211](https://github.com/joshuar/go-hass-agent/commit/4fba21186cac1a3aa47e11cd78d0db7717999ca1))
* **agent:** rework sensor table to use Fyne table widget ([33423c5](https://github.com/joshuar/go-hass-agent/commit/33423c54b07c2a95e8df30dc6e6366b7e34cd774))
* **hass:** fix error message ([ef1d239](https://github.com/joshuar/go-hass-agent/commit/ef1d239b26afad9c41c8a606a97e041bd715123a))
* **linux:** remove attributes for active app sensor to avoid memory leak ([e700db5](https://github.com/joshuar/go-hass-agent/commit/e700db5e501a9bad9538c7d9c14a1fe75a6c8d74))

## [1.2.1](https://github.com/joshuar/go-hass-agent/compare/v1.2.0...v1.2.1) (2023-06-11)


### Bug Fixes

* **agent:** no need to wait on app quit ([b59a31e](https://github.com/joshuar/go-hass-agent/commit/b59a31e21c4e37db594fd3be8136f7838c4d5e1d))

## [1.2.0](https://github.com/joshuar/go-hass-agent/compare/v1.1.0...v1.2.0) (2023-06-11)


### Features

* **assets:** add systemd service file ([1d6f695](https://github.com/joshuar/go-hass-agent/commit/1d6f695249c9b22999ae9f5dc4b5363a2f3b4003))


### Bug Fixes

* **hass:** remove unused code in websockets ([3d9f466](https://github.com/joshuar/go-hass-agent/commit/3d9f4668c38e44e5538072ad172dc4252c413c2f))
* **linux:** catch error ([fd60fc4](https://github.com/joshuar/go-hass-agent/commit/fd60fc42e4ccb6c101674bc3dc341dffc3e2ebd9))

## [1.1.0](https://github.com/joshuar/go-hass-agent/compare/v1.0.1...v1.1.0) (2023-06-06)


### Features

* **linux:** add screen lock sensor ([ffb7276](https://github.com/joshuar/go-hass-agent/commit/ffb72768085176ff06026469f85a885a2debfece))


### Bug Fixes

* **agent:** correct formatting of debug messages ([33346c1](https://github.com/joshuar/go-hass-agent/commit/33346c12b782d9aacb1dbcc07d1f8167c005487c))
* **hass:** add omitempty to JSON fields where appropriate ([a139511](https://github.com/joshuar/go-hass-agent/commit/a139511b19dd8ce6f4abbbbf1c32526bb5aa4058))
* **hass:** update API tests ([a1daa09](https://github.com/joshuar/go-hass-agent/commit/a1daa091e7601c9b6c22198ae31ae8f436cd843f))
* **linux:** remove unneeded parameters ([42af00d](https://github.com/joshuar/go-hass-agent/commit/42af00d675e50d9375a73174f236813fe06fdc05))

## [1.0.1](https://github.com/joshuar/go-hass-agent/compare/v1.0.0...v1.0.1) (2023-06-03)


### Bug Fixes

* **agent:** handle no sensors data display ([4d22c2f](https://github.com/joshuar/go-hass-agent/commit/4d22c2f7df90880c0d4599ac7e057862418ca8ab))
* **agent:** use a waitgroup for registration ([3461859](https://github.com/joshuar/go-hass-agent/commit/3461859bc0c48e3a89d8c658232cf0e1feb806f0))
* **hass:** handle empty config response ([f4b4412](https://github.com/joshuar/go-hass-agent/commit/f4b4412bebae1ce423b874de3a996911ffc3da21))
* **hass:** omit sending response body if empty or nil ([7281cbb](https://github.com/joshuar/go-hass-agent/commit/7281cbbd9de1e3d50b34f08ce288928199fd2618))
* **sensors,hass:** merge Sensor and SensorUpdate interfaces ([aa9e200](https://github.com/joshuar/go-hass-agent/commit/aa9e200558bffe07610c65dbb2220830d1840564))

## [1.0.0](https://github.com/joshuar/go-hass-agent/compare/v0.4.0...v1.0.0) (2023-05-20)


### ⚠ BREAKING CHANGES

* **linux:** remove unused context management for previous api
* **linux:** remove deprecated functions and rework api struct
* **agent,linux:** utilise the device API interface in agent code
* **device,linux,sensors:** remove sensorinfo struct, move sensor workers into device API

### Features

* **agent:** add a (very rough) window to display all sensors and their states ([f7fb4b9](https://github.com/joshuar/go-hass-agent/commit/f7fb4b9c13cdc37757a768ade9cdca853bf6e774))
* **agent:** add tray menu option to access fyne settings ([3ef943e](https://github.com/joshuar/go-hass-agent/commit/3ef943e19342645d3794a0407cfd8e1a5be4721a))
* **agent:** report Home Assistant version in About dialog ([a7e1d83](https://github.com/joshuar/go-hass-agent/commit/a7e1d8311fe07ba8b9353ffa8cfb9a1ee9148697))


### Bug Fixes

* **agent:** better initial window size for sensors display ([c46792f](https://github.com/joshuar/go-hass-agent/commit/c46792f735a657c1e6e72a623fc46091c9361c5b))
* **agent:** correct config validation returns ([8c89544](https://github.com/joshuar/go-hass-agent/commit/8c89544f95e68e7921a97deb590810e87f3ddb16))
* **agent:** waitgroup decrement for worker finish ([42ff9f0](https://github.com/joshuar/go-hass-agent/commit/42ff9f03f71aa52ddf7791d1b13d702b692aceca))
* **device,linux:** create and use a safer function for getting an endpoint from the API interface ([d5a8c66](https://github.com/joshuar/go-hass-agent/commit/d5a8c666d23e8c88a4fa20cb5800ace06e2cd662))
* **device:** use a context with timeout for fetching external ip ([a888d5f](https://github.com/joshuar/go-hass-agent/commit/a888d5fee30ff91b5708f21ca0aecf404eaffc4c))
* **hass:** cancel websocket connection context when done message received ([6ecfc8e](https://github.com/joshuar/go-hass-agent/commit/6ecfc8ea21075a313278a323dfba7e8b233658bd))
* **hass:** don't run response handler if request was never sent ([dc6e4b1](https://github.com/joshuar/go-hass-agent/commit/dc6e4b189019cc05ed1291512c604eb7e51d6194))
* **hass:** make sure api context is cancelled/closed in all branches ([f40971d](https://github.com/joshuar/go-hass-agent/commit/f40971d72e4f70fe49ae4035417c7f7462655ce0))
* **linux:** bail early if the matched signal doesn't have a body ([d53eb02](https://github.com/joshuar/go-hass-agent/commit/d53eb0209078260a9919ad9ead43870b0d2d8cd8))
* **linux:** clean up finding processes ([8856bb2](https://github.com/joshuar/go-hass-agent/commit/8856bb2ced6d71cc4731c9ee5b38721bee99c15e))
* **linux:** remove outdated external package for geoclue/location ([22a4b7f](https://github.com/joshuar/go-hass-agent/commit/22a4b7ff065bc862df9b092c04f1f7ccaba48f4b))
* **linux:** safer access to api endpoint map ([f7defa3](https://github.com/joshuar/go-hass-agent/commit/f7defa3bedfcc5b3af7782e97df33e4c369ac539))
* **sensors:** bail on error getting sensor workers ([1f1442b](https://github.com/joshuar/go-hass-agent/commit/1f1442b954c9a156b64a00462e8698d383cc2c52))


### Code Refactoring

* **agent,linux:** utilise the device API interface in agent code ([a87cc06](https://github.com/joshuar/go-hass-agent/commit/a87cc06b1f554e2f52e055180be020fa1857f138))
* **device,linux,sensors:** remove sensorinfo struct, move sensor workers into device API ([bceea60](https://github.com/joshuar/go-hass-agent/commit/bceea60df97fbf952966d8a697336373e3660c86))
* **linux:** remove deprecated functions and rework api struct ([6125e9f](https://github.com/joshuar/go-hass-agent/commit/6125e9faa8e51d7c6512adb603a4710b54434e0a))
* **linux:** remove unused context management for previous api ([dd740e5](https://github.com/joshuar/go-hass-agent/commit/dd740e5c5cc2b8ba4f0217958171cd3a598703af))

## [0.4.0](https://github.com/joshuar/go-hass-agent/compare/v0.3.5...v0.4.0) (2023-05-14)


### Features

* add a way to run "headless" (without any GUI) ([90b9a82](https://github.com/joshuar/go-hass-agent/commit/90b9a82eb5baab2b95f24187b5f9996a8c0b4fbc))
* **cmd:** add a "register" command ([0fa27ad](https://github.com/joshuar/go-hass-agent/commit/0fa27ade51024f41e3977c0d838dc46bf0ada30e))
* **hass:** add backoff functionality for registration requests ([d268cff](https://github.com/joshuar/go-hass-agent/commit/d268cffc9d0695b20ed3c93acc148b154d1dd49d))


### Bug Fixes

* **agent:** ensure preferences get saved ([30e2735](https://github.com/joshuar/go-hass-agent/commit/30e273568330a1429ab593a0f93c865529bf687c))
* **hass:** fix backoff package dependency ([9b8f519](https://github.com/joshuar/go-hass-agent/commit/9b8f5198a76cf051734d534ac4d56ac29758dec4))
* **hass:** id increment for websocket requests ([f941fb6](https://github.com/joshuar/go-hass-agent/commit/f941fb6caff929937a4a69f6563c045248258676))
* **hass:** improve websocket resiliency with ping/pong logic ([6428b77](https://github.com/joshuar/go-hass-agent/commit/6428b77fb146b0062ab0d9b2f2eccd1545d70534))

## [0.3.5](https://github.com/joshuar/go-hass-agent/compare/v0.3.0...v0.3.5) (2023-05-11)


### Features

* **agent:** add "report issue" and "request feature" actions to tray icon menu ([7f75a52](https://github.com/joshuar/go-hass-agent/commit/7f75a5232d8d2f611c2152a81dd9f687643dfd45))


### Bug Fixes

* **hass:** fix websocket not working and migrate to different websocket package ([2607346](https://github.com/joshuar/go-hass-agent/commit/26073467734bc81e5d567099f91de0ba70ad7f6c))
* **linux:** uptime sensor now measured in hours ([f8fccd2](https://github.com/joshuar/go-hass-agent/commit/f8fccd275fcb1fb2bc9e4105295a22d83eef4869))


### Miscellaneous Chores

* release 0.3.5 ([b613696](https://github.com/joshuar/go-hass-agent/commit/b613696729bd4d858f3121ceef7ab72385437702))

## [0.3.0](https://github.com/joshuar/go-hass-agent/compare/v0.2.0...v0.3.0) (2023-05-07)


### Features

* **linux:** add uptime and last boot sensors ([1b5cbb1](https://github.com/joshuar/go-hass-agent/commit/1b5cbb1d70107f8fb74badb95442b0c542262c2d))
* **linux:** improve network connection sensor code ([5a18911](https://github.com/joshuar/go-hass-agent/commit/5a18911dec3cb664797830047093d94d51d02ed3))
* **linux:** move to a single dbus watch for networkmanager ([89569dd](https://github.com/joshuar/go-hass-agent/commit/89569dd75e2df2b732de9cb7557f47ae7c140f35))


### Bug Fixes

* **build:** fix dependencies for rpm/deb packages ([0a16bd2](https://github.com/joshuar/go-hass-agent/commit/0a16bd2cc86ca1ac26edb82e978d4de5bd37cbbc))
* **linux:** remove spew ([c77c4be](https://github.com/joshuar/go-hass-agent/commit/c77c4beeab964a4fb849cc41bdc7129e989fed30))

## [0.2.0](https://github.com/joshuar/go-hass-agent/compare/v0.1.0...v0.2.0) (2023-05-02)


### Features

* better handling of app interrupt/termination ([47dce34](https://github.com/joshuar/go-hass-agent/commit/47dce3427ce2b0607b9ab1654333826f05191313))
* **sensors:** add a nutsDB registry backend ([6a7cb84](https://github.com/joshuar/go-hass-agent/commit/6a7cb84a9423178d81f91906d7bef1d6a2384c86))
* **sensors:** implement a Registry interface ([80c4c0a](https://github.com/joshuar/go-hass-agent/commit/80c4c0abdcdb4da8a2459ac495eb49d47a8fd28b))
* **sensors:** use nutsDB registry backend by default ([e533f91](https://github.com/joshuar/go-hass-agent/commit/e533f917397a2718a92e2e39f5eef8abf3fdcfe9))


### Bug Fixes

* **agent:** remove spew debug message ([0b8e5bf](https://github.com/joshuar/go-hass-agent/commit/0b8e5bffc279996933c80e5f15ca53413287e75d))


### Miscellaneous Chores

* release 0.1.1 ([84526ab](https://github.com/joshuar/go-hass-agent/commit/84526ab7001ecdd3f00916828f55926afbdbb5aa))
* release 0.2.0 ([118dbc4](https://github.com/joshuar/go-hass-agent/commit/118dbc45e6fb5e96f28d75ce86d7851f110f4d2a))

## [0.1.0](https://github.com/joshuar/go-hass-agent/compare/v0.0.7...v0.1.0) (2023-04-27)


### Features

* **device_linux:** add disk usage sensors ([d85d79e](https://github.com/joshuar/go-hass-agent/commit/d85d79e5377a00a201326ac253ac6e75716d6a66))
* **linux:** add network bytes received/sent sensors ([a71effc](https://github.com/joshuar/go-hass-agent/commit/a71effc754cce83e965ff7ccac0d6eae4be518cc))


### Bug Fixes

* command-line description FR working this time... ([d89c265](https://github.com/joshuar/go-hass-agent/commit/d89c2652cdc17c66c33e1c5993e7ce8d5591f4f6))
* **hass,device:** fix incorrect state/device class types ([85abcd8](https://github.com/joshuar/go-hass-agent/commit/85abcd8d76c6d6a6d63e64bb9e684bc640007876))
* **linux:** add units to sensors ([97db5cf](https://github.com/joshuar/go-hass-agent/commit/97db5cf8e000ed4c4b74d62684d5ba92f4f61ff6))
* missing app description for command-line! ([6f49f22](https://github.com/joshuar/go-hass-agent/commit/6f49f22cd885e04c1d0d1428eca903605aff0ac9))

## [0.0.7](https://github.com/joshuar/go-hass-agent/compare/v0.0.6...v0.0.7) (2023-04-22)


### Features

* add a problems sensor on Linux to track problems reported by abrt ([59978d6](https://github.com/joshuar/go-hass-agent/commit/59978d698c2301981369fa265e099a85e9db3b80))
* add memory usage sensor on Linux ([81b9a2e](https://github.com/joshuar/go-hass-agent/commit/81b9a2e21d58b3241eb6df364dd0916a60b01e11))
* add swap memory sensors on Linux ([6dfd177](https://github.com/joshuar/go-hass-agent/commit/6dfd177a917589ffc45742da622fabd8a1d4b255))
* and active command start time to active apps sensor. ([992d832](https://github.com/joshuar/go-hass-agent/commit/992d83250f3c3e8342aa781b62f3e11dc754afd4))
* use jitterbug package to add some jitter to polling sensors to ([85366e7](https://github.com/joshuar/go-hass-agent/commit/85366e7b40c0728f9bcfe708a59a69a300f4d78d))


### Bug Fixes

* GetDBusData* functions properly handle optional arguments. ([992d832](https://github.com/joshuar/go-hass-agent/commit/992d83250f3c3e8342aa781b62f3e11dc754afd4))
* handle unknown process creation time attribute ([73b3904](https://github.com/joshuar/go-hass-agent/commit/73b39041d03022cbb3a6d71417e4ce7b96fd208c))
* memory and loadavg sensors work properly on Linux now ([5b7e922](https://github.com/joshuar/go-hass-agent/commit/5b7e922ed762569f153e4c0423e9d234596c28fa))
* naming of variable for release please ([bfc4336](https://github.com/joshuar/go-hass-agent/commit/bfc433661ab478d1f42c0646989abac4916acdd6))
* release please use token ([d6cd37f](https://github.com/joshuar/go-hass-agent/commit/d6cd37f556527df0ebced398016ddf5aef16413f))


### Miscellaneous Chores

* release 0.0.7 ([eb143f6](https://github.com/joshuar/go-hass-agent/commit/eb143f6ebf1eaf06d7d69afb7e95dc6376667cda))
* release 0.0.7 ([0b5973f](https://github.com/joshuar/go-hass-agent/commit/0b5973fcaabf045fc32a04c040c6acde20a56e38))
