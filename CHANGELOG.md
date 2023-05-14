# Changelog

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
