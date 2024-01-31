// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

const (
	PrefMQTTServerHelp   = "Format should be scheme://host:port Where 'scheme' is one of 'tcp' or 'ssl', 'host' is the ip-address (or hostname) and 'port' is the port on which the broker is accepting connections."
	PrefMQTTUserHelp     = "Optional username to authenticate with the broker."
	PrefMQTTPasswordHelp = "Optional password to authenticate with the broker."
)
