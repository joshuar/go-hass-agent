# Extending the Agent

## Adding OS support

- The intention of the agent design is to make it OS-agnostic.
- Most OS specific code for fetching sensor data should likely be part of the
  device package and using filename suffixes such as `filename_GOOS_GOARCH.go`. 
- You'll need to implement the functionality required for each of
  the worker functions for getting sensor data the agent runs, that pull/push
  data to Home Assistant. Specific details are further in this document.
- For some OSes, you might need some code to initialise or create some data source
  or API that the individual sensor fetching code uses. 
  - For example, on Linux, a DBus connection is used for a lot of the sensor data gathering.
- In such cases, you should pass this through as a value in a context. This
  context should be created in the `agent.Run()` function, before the agent
  itself is created with `agent.New(ctx)`. You can then pass your context
  initialised with whatever you need to the latter and it will propagate through
  the agent code.
- Then, when you need access to your data/APIs, just fetch it from the context
  value as needed. 

### Worker Implementation Details

#### Location Information

- Implement a function `LocationUpdater(ctx, chan interface{})` that can be run in a goroutine.
- The function should use the passed channel to send location updates when
  needed. The data sent should satisfy the `LocationInfo` interface in
  `internal/agent/location.go`.
- The function should take the passed context and derive its own where needed.
  It should handle the context being cancelled and gracefully stop work.

#### App Sensors

- Create a function `AppUpdater(ctx, chan interface{})` that can be run in a goroutine.
- The function should send data on the channel that that implements both the `activeApp` and `runningApps`
  interfaces in `internal/agent/sensorApps.go`. 
- The function should take the passed context and derive its own where needed.
  It should handle the context being cancelled and gracefully stop work.
