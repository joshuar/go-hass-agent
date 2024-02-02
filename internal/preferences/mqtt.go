// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

// MQTTPrefereneces encapsulates Preferences so it can be passed to an MQTT
// client and satisfy the config interface that code requires.
type MQTTPreferences struct {
	Prefs *Preferences
}

// MQTTEnabled returns whether MQTT is enabled in the agent.
func (p *MQTTPreferences) MQTTEnabled() bool {
	return p.Prefs.MQTTEnabled
}

// MQTTServer returns the broker URI from the preferences.
func (p *MQTTPreferences) MQTTServer() string {
	return p.Prefs.MQTTServer
}

// MQTTUser returns any username required for connecting to the broker from the
// preferences.
func (p *MQTTPreferences) MQTTUser() string {
	return p.Prefs.MQTTUser
}

// MQTTPassword returns any password required for connecting to the broker from the
// preferences.
func (p *MQTTPreferences) MQTTPassword() string {
	return p.Prefs.MQTTPassword
}
