// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

type MQTTPreferences struct {
	Prefs *Preferences
}

func (p *MQTTPreferences) MQTTEnabled() bool {
	return p.Prefs.MQTTEnabled
}

func (p *MQTTPreferences) MQTTServer() string {
	return p.Prefs.MQTTServer
}

func (p *MQTTPreferences) MQTTUser() string {
	return p.Prefs.MQTTUser
}

func (p *MQTTPreferences) MQTTPassword() string {
	return p.Prefs.MQTTPassword
}
