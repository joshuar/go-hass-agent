// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:lll
package ui

const (
	RegistrationInfoString = `To register the agent, please enter the relevant details for your Home Assistant
server (if not auto-detected) and long-lived access token.`

	InvalidURLMsgString      = `You need to specify a valid http(s)://host:port.`
	InvalidURIMsgString      = `You need to specify a valid scheme://host:port.`
	InvalidHostPortMsgString = `You need to specify a valid host:port combination.`

	MQTTServerInfoString   = "Format should be scheme://host:port Where 'scheme' is one of 'tcp' or 'ssl', 'host' is the ip-address (or hostname) and 'port' is the port on which the broker is accepting connections."
	MQTTUserInfoString     = "Optional username to authenticate with the broker."
	MQTTPasswordInfoString = "Optional password to authenticate with the broker."

	PrefsRestartMsgString = `Please restart the agent to use changed preferences.`
)
