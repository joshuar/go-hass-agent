# Extending the Agent

## Adding OS support

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
