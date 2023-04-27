# Changelog

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
