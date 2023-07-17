// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

//go:generate moq -out mock_Config_test.go . Config
type Config interface {
	WebSocketURL() string
	WebhookID() string
	Token() string
	ApiURL() string
	Secret() string
}
