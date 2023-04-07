package device

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

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
}

func GetDeviceInfo(d DeviceInfo) {
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

func GenerateRegistrationRequest(d DeviceInfo) *hass.RegistrationRequest {
	if d.AppData() != nil {
		return &hass.RegistrationRequest{
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
		return &hass.RegistrationRequest{
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

type SensorInfo struct {
	sensorWorkers map[string]func(context.Context, chan interface{}, chan struct{})
}

func NewSensorInfo() *SensorInfo {
	return &SensorInfo{
		sensorWorkers: make(map[string]func(context.Context, chan interface{}, chan struct{})),
	}
}

func (i *SensorInfo) Add(name string, workerFunc func(context.Context, chan interface{}, chan struct{})) {
	log.Debug().Caller().
		Msgf("Added a sensorWorker for %s", name)
	i.sensorWorkers[name] = workerFunc
}

func (i *SensorInfo) Get() map[string]func(context.Context, chan interface{}, chan struct{}) {
	return i.sensorWorkers
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for device.deviceAPI values in Contexts. It is
// unexported; clients use device.NewContext and device.FromContext
// instead of using this key directly.
var configKey key

// NewContext returns a new Context that carries value d.
func NewContext(ctx context.Context, d *deviceAPI) context.Context {
	return context.WithValue(ctx, configKey, d)
}

// FromContext returns the deviceAPI value stored in ctx, if any.
func FromContext(ctx context.Context) (*deviceAPI, bool) {
	c, ok := ctx.Value(configKey).(*deviceAPI)
	return c, ok
}
