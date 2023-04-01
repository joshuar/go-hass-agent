# Agent
## Configuration

- The agent needs a valid configuration and will prompt the user to help create it if one isn't found.
- Configuration is stored using the [Fyne Preferences API](https://developer.fyne.io/explore/preferences). 
- `LoadConfig` in `internal/agent/config.go` is responsible for loading the
  config. It will also create a new config (prompting the user for required
  details via `runRegistrationWorker`) if needed. 


## Workers

- Most of the agent work is done by worker functions, named `runSomethingWorker`.  
  - For example, to track and update location, there is `runLocationWorker` in `internal/agent/location.go`.
- Once the config has been loaded, all the worker functions are started in separate goroutines.

### Contexts

The agent creates a cancellable context for itself. Any worker goroutines should
derive their own contexts from this (adding their own timeouts, cancellations as
needed). 

## Data Updates

- Data updates are left entirely up to the individual sensors and the platform
  the app is running on. For example, on Linux, location and running apps are
  updated as often as the relevant information is published on the user's
  session DBus. 
- Sensors push data to the app's worker routines which handle
  publishing it to Home Assistant. 
## Notifications

- The agent supports receiving notifications from Home Assistant via a
  [websocket
  connection](https://developers.home-assistant.io/docs/api/native-app-integration/notifications#enabling-websocket-push-notifications).
- Notifications are displayed using the [Fyne Notification
  API](https://developer.fyne.io/api/v2.3/notification.html). 
