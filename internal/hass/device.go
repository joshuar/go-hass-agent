package hass

import (
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/rs/zerolog/log"
)

type deviceInfo interface {
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
}

func NewDevice() deviceInfo {
	return device.NewDevice()
}

func GetDeviceInfo(d deviceInfo) {
	log.Info().Msgf("Device ID: %s", d.DeviceID())
	log.Info().Msgf("Device Name: %s", d.DeviceName())
	log.Info().Msgf("App ID: %s", d.AppID())
	log.Info().Msgf("App Name: %s", d.AppName())
	log.Info().Msgf("App Verson: %s", d.AppVersion())
	log.Info().Msgf("Manufacturer: %s", d.Manufacturer())
	log.Info().Msgf("Model: %s", d.Model())
	log.Info().Msgf("OS Name: %s", d.OsName())
	log.Info().Msgf("OS Version: %s", d.OsVersion())
	log.Info().Msgf("Supports Encryption: %v", d.SupportsEncryption())
}
