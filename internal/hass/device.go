package hass

import (
	"runtime"

	"github.com/joshuar/go-hass-agent/internal/linux"
	log "github.com/sirupsen/logrus"
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
	switch os := runtime.GOOS; os {
	case "linux":
		return linux.NewLinuxDevice()
	default:
		log.Error("Unsupported Operating System.")
		return nil
	}
}

func GetDeviceInfo(d deviceInfo) {
	log.Infof("Device ID: %s", d.DeviceID())
	log.Infof("Device Name: %s", d.DeviceName())
	log.Infof("App ID: %s", d.AppID())
	log.Infof("App Name: %s", d.AppName())
	log.Infof("App Verson: %s", d.AppVersion())
	log.Infof("Manufacturer: %s", d.Manufacturer())
	log.Infof("Model: %s", d.Model())
	log.Infof("OS Name: %s", d.OsName())
	log.Infof("OS Version: %s", d.OsVersion())
	log.Infof("Supports Encryption: %v", d.SupportsEncryption())
}

func GenerateRegistrationRequest(d deviceInfo) *RegistrationRequest {
	if d.AppData() != nil {
		return &RegistrationRequest{
			DeviceID:           d.DeviceID(),
			AppID:              d.AppID(),
			AppName:            d.AppName(),
			AppVersion:         d.AppVersion(),
			DeviceName:         d.DeviceName(),
			Manufacturer:       d.Manufacturer(),
			Model:              d.Model(),
			OsName:             d.OsName(),
			OsVersion:          d.OsVersion(),
			SupportsEncryption: d.SupportsEncryption(),
			AppData:            d.AppData(),
		}
	} else {
		return &RegistrationRequest{
			DeviceID:           d.DeviceID(),
			AppID:              d.AppID(),
			AppName:            d.AppName(),
			AppVersion:         d.AppVersion(),
			DeviceName:         d.DeviceName(),
			Manufacturer:       d.Manufacturer(),
			Model:              d.Model(),
			OsName:             d.OsName(),
			OsVersion:          d.OsVersion(),
			SupportsEncryption: d.SupportsEncryption(),
		}
	}
}
