<!--
 Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# go-hass-agent Development

## Build Requirements

`go-hass-agent` has a number of build requirements external to `go.mod` that
need to be installed:

- [stringer](https://pkg.go.dev/golang.org/x/tools/cmd/stringer).
  - Typically installed with `go install golang.org/x/tools/cmd/stringer@latest`.
- [go-enum](https://github.com/abice/go-enum).
  - Typically installed with `go install github.com/abice/go-enum@latest`.
- [fyne-cross](https://github.com/fyne-io/fyne-cross).
  - Typically installed with `go install github.com/fyne-io/fyne-cross@latest`
- [gotext](https://cs.opensource.google/go/x/text)
  - Typically installed with `go install golang.org/x/text/cmd/gotext@latest`
- [goreleaser](https://goreleaser.com/install/).

## Extending the Agent

### Adding OS support

The intention of the agent design is to make it OS-agnostic.

Most OS specific code for fetching sensor data should likely be part of the
`device` package and using filename suffixes such as `filename_GOOS_GOARCH.go`. 

For some OSes, you might need some code to initialise or create some data source
or API that the individual sensor fetching code uses. 

For example, on Linux, a DBus connection is used for a lot of the sensor data gathering.

In such cases, you should pass this through as a value in a context. You can
create the following function for your platform:

```go
SetupContext(ctx context.Context) context.Context
```

It should accept a `context.Context` and derive its own context from this base
that contains the necessary values for the platform. It will be propagated
throughout the code wherever a context is passed and available for retrieval and
use.

An example can be found in `device/helpers_linux.go`.

### Adding sensors

See [device/sensors.md](device/sensors.md).

## Committing Code

This repository is using [conventional commit
messages](https://www.conventionalcommits.org/en/v1.0.0/#summary). This provides
the ability to automatically include relevant notes in the
[changelog](../CHANGELOG.md). The [TL;DR](https://en.wikipedia.org/wiki/TL;DR)
is when writing commit messages, add a prefix:

- `feat:` for a new feature, like a new sensor.
- `fix:` when fixing an issue.
- `refactor:` when making non-visible but useful code changes.
- …and so on. See the link above or see the existing commit messages for examples.
