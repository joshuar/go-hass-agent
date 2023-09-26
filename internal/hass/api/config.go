// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

//go:generate moq -out mock_Agent_test.go . Agent
type Agent interface {
	GetConfig(string, interface{}) error
}
