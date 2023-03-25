# Agent

## Architecture

### Configuration

- The agent needs a valid configuration and will prompt the user to help create it if one isn't found.
- Configuration is stored using the [Fyne Preferences API](https://developer.fyne.io/explore/preferences). 
- `LoadConfig` in `internal/agent/config.go` is responsible for loading the
  config. It will also create a new config (prompting the user for required
  details via `runRegistrationWorker`) if needed. 

### Workers

- Most of the agent work is done by worker functions, named `runSomethingWorker`.  
  - For example, to track and update location, there is `runLocationWorker` in `internal/agent/location.go`.
- Once the config has been loaded, all the worker functions are started in separate goroutines.


