// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

type config interface {
	WebSocketURL() string
	WebhookID() string
	Token() string
}
