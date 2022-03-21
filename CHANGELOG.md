## [2.5.1](https://github.com/datastack-net/dockerized/compare/v2.5.0...v2.5.1) (2022-03-21)


### Bug Fixes

* building commands outside dockerized working directory ([6cd4701](https://github.com/datastack-net/dockerized/commit/6cd4701d37c53903f2827b0935163179bce2cf05))

# [2.5.0](https://github.com/datastack-net/dockerized/compare/v2.4.0...v2.5.0) (2022-03-20)


### Features

* Downloads for Linux, Mac and Windows ([7a5d4a6](https://github.com/datastack-net/dockerized/commit/7a5d4a6ec1729b42c5e03e9bcd97e8a9def06294))

# [2.4.0](https://github.com/datastack-net/dockerized/compare/v2.3.0...v2.4.0) (2022-03-20)


### Bug Fixes

* allow overriding GOARCH for 'go' when running from source ([fd34712](https://github.com/datastack-net/dockerized/commit/fd347121dd53ad6195c9655dcb32da64d35b72b6))


### Features

* pass GOARCH to go ([0934e53](https://github.com/datastack-net/dockerized/commit/0934e5355653e68f3ae6599cfb5d78705f5671b5))

# [2.3.0](https://github.com/datastack-net/dockerized/compare/v2.2.4...v2.3.0) (2022-03-20)


### Bug Fixes

* --shell for alpine-based commands ([e3c3e86](https://github.com/datastack-net/dockerized/commit/e3c3e86ecaf4e5d4533ef3a36b556eaf7de0fc1f))


### Features

* add 'rust', 'zip' ([07af7d3](https://github.com/datastack-net/dockerized/commit/07af7d3856473e686e262be467008089be83e7b6))

## [2.2.4](https://github.com/datastack-net/dockerized/compare/v2.2.3...v2.2.4) (2022-03-20)


### Bug Fixes

* include .env file in windows build ([5f7c435](https://github.com/datastack-net/dockerized/commit/5f7c435ce2ed455e28ae983cef4db485a3346299))

## [2.2.3](https://github.com/datastack-net/dockerized/compare/v2.2.2...v2.2.3) (2022-03-20)


### Bug Fixes

* dockerized overrides GOOS env var, breaking windows build on fresh systems ([21703f6](https://github.com/datastack-net/dockerized/commit/21703f60f2ffe96131f1022884bfa58c59981c33))

## [2.2.2](https://github.com/datastack-net/dockerized/compare/v2.2.1...v2.2.2) (2022-03-20)


### Bug Fixes

* Windows build not executable on Windows ([528c7d8](https://github.com/datastack-net/dockerized/commit/528c7d83980bb88337245a2ff1dd684c9c3f7366))

## [2.2.1](https://github.com/datastack-net/dockerized/compare/v2.2.0...v2.2.1) (2022-03-20)


### Bug Fixes

* pre-compiled Windows binary could not run because of missing compose file ([e0fab23](https://github.com/datastack-net/dockerized/commit/e0fab23fe3c9d8f8baccd948bc172c88e15f3482))
* pre-compiled Windows binary could not run because of missing compose file ([48cbe6a](https://github.com/datastack-net/dockerized/commit/48cbe6a442b501078f0d1227344f86edb145aca3))
* remove accidentally added services that triggered warnings ([81dc609](https://github.com/datastack-net/dockerized/commit/81dc609d3f478994542470f42744ee0ee0eac655))

# [2.2.0](https://github.com/datastack-net/dockerized/compare/v2.1.0...v2.2.0) (2022-03-20)


### Features

* add 'zip' ([82a753c](https://github.com/datastack-net/dockerized/commit/82a753cee8470bffe2c98707b3bb8f70240a5b39))
* automatically release windows binaries ([3c73176](https://github.com/datastack-net/dockerized/commit/3c73176840d9127c0d9d96e316040964dd6d7ad2))
* Pass GOOS env to 'go' command for cross-compilation ([416d14b](https://github.com/datastack-net/dockerized/commit/416d14b6baf8e57d3c80186738a8a74c5262cca5))

# [2.1.0](https://github.com/datastack-net/dockerized/compare/v2.0.0...v2.1.0) (2022-03-20)


### Bug Fixes

* running --shell without arguments ([e9573ee](https://github.com/datastack-net/dockerized/commit/e9573eedc4ce20fb1da8c40f1969f0b81ad7b2ca))


### Features

* automatic releases ([ac2629d](https://github.com/datastack-net/dockerized/commit/ac2629da96a7197fdddc79320d75d7db120bae2e))