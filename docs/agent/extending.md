# Extending the Agent

## Adding OS support

- The intention of the agent design is to make it OS-agnostic.
- All OS specific code should go under `internal/sensors` and using filename
  suffixes such as `filename_GOOS_GOARCH.go`. 
- In general, you'll need to implement the functionality required for each of
  the worker functions the agent runs, that pull/push data to Home Assistant.
  Specific details are further in this document.


### Worker Implementation Details

#### Location Information

- Implement a function `LocationUpdater` that can be run in a goroutine, that accepts a parameter
  of type `chan interface{}`.
- The function should use the passed channel to send location updates when
  needed. The data sent should satisfy the `LocationInfo` interface in
  `internal/agent/location.go`.


