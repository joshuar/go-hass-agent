# Changelog

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
