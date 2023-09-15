// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

//go:generate moq -out mock_DeviceInfo_test.go . DeviceInfo
type DeviceInfo interface {
	DeviceID() string
	AppID() string
	AppName() string
	AppVersion() string
	DeviceName() string
	Manufacturer() string
	Model() string
	OsName() string
	OsVersion() string
	SupportsEncryption() bool
	AppData() interface{}
	MarshalJSON() ([]byte, error)
}
