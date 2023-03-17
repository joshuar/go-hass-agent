package hass

import (
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
	AppData() *AppData
}

type AppData struct {
	PushNotificationKey string `json:"push_notification_key"`
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

