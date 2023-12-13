# Changelog

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
