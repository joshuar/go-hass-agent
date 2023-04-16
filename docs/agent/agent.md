<!--
 Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Agent
## Configuration

- The agent needs a valid configuration and will prompt the user to help create it if one isn't found.
- Configuration is stored using the [Fyne Preferences API](https://developer.fyne.io/explore/preferences). 
- `LoadConfig` in `internal/agent/config.go` is responsible for loading the
  config. It will also create a new config (prompting the user for required
  details via `runRegistrationWorker`) if needed. 

### Contexts

The agent creates a cancellable context for itself. Any platform code should
accept this context and handle context cancellation gracefully (adding their own
timeouts, cancellations as needed). More details in [extending](extending.md).

## Data Updates

- Data updates are left entirely up to the individual sensors and the platform
  the app is running on. For example, on Linux, location and running apps are
  updated as often as the relevant information is published on the user's
  session DBus. More details about how sensor updates work can be found in
  [device/sensors](../device/sensors.md).
## Notifications

- The agent supports receiving notifications from Home Assistant via a
  [websocket
  connection](https://developers.home-assistant.io/docs/api/native-app-integration/notifications#enabling-websocket-push-notifications).
- Notifications are displayed using the [Fyne Notification
  API](https://developer.fyne.io/api/v2.3/notification.html). 
