// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

type Hass struct {
	CloudhookURL   string `toml:"cloudhook_url,omitempty" json:"cloudhook_url" validate:"omitempty,http_url"`
	RemoteUIURL    string `toml:"remote_ui_url,omitempty" json:"remote_ui_url" validate:"omitempty,http_url"`
	Secret         string `toml:"secret,omitempty" json:"secret" validate:"omitempty,ascii"`
	WebhookID      string `toml:"webhook_id" json:"webhook_id" validate:"required,ascii"`
	RestAPIURL     string `toml:"apiurl,omitempty" json:"-" validate:"required_without=CloudhookURL RemoteUIURL,http_url"`
	WebsocketURL   string `toml:"websocketurl" json:"-" validate:"required,url"`
	IgnoreHassURLs bool   `toml:"ignore_hass_urls,omitempty" json:"-" validate:"omitempty,boolean"`
}
