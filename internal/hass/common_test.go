// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"encoding/json"

	"github.com/stretchr/testify/mock"
)

type mockRequest struct {
	mock.Mock
}

func (m *mockRequest) RequestType() RequestType {
	args := m.Called()
	return args.Get(0).(RequestType)
}

func (m *mockRequest) RequestData() json.RawMessage {
	args := m.Called()
	return args.Get(0).(json.RawMessage)
}

func (m *mockRequest) ResponseHandler(b bytes.Buffer) {
	m.Called(b)
}

type mockConfig struct {
	mock.Mock
}

func (m *mockConfig) Get(property string) (interface{}, error) {
	args := m.Called(property)
	return args.String(0), args.Error(1)
}

func (m *mockConfig) Set(property string, value interface{}) error {
	args := m.Called(property, value)
	return args.Error(1)
}

func (m *mockConfig) Validate() error {
	args := m.Called()
	return args.Error(1)
}

func (m *mockConfig) Upgrade() error {
	args := m.Called()
	return args.Error(1)
}

func (m *mockConfig) Refresh() error {
	args := m.Called()
	return args.Error(1)
}

type mockDevice struct {
	mock.Mock
}

func (d *mockDevice) DeviceID() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) AppID() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) AppName() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) AppVersion() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) DeviceName() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) Manufacturer() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) Model() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) OsName() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) OsVersion() string {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) SupportsEncryption() bool {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) AppData() interface{} {
	panic("not implemented") // TODO: Implement
}

func (d *mockDevice) MarshalJSON() ([]byte, error) {
	panic("not implemented") // TODO: Implement
}
