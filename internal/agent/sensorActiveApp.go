package agent

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type activeApp interface {
	Name() string
}

type activeAppSensor struct {
	state      string
	dataCh     chan interface{}
	disabled   bool
	registered bool
}

func (s *activeAppSensor) Attributes() interface{} {
	return nil
}

func (s *activeAppSensor) DeviceClass() string {
	return ""
}

func (s *activeAppSensor) Icon() string {
	return "mdi:application"
}

func (s *activeAppSensor) Name() string {
	return "Active Application"
}

func (s *activeAppSensor) State() interface{} {
	return s.state
}

func (s *activeAppSensor) SensorType() string {
	return "sensor"
}

func (s *activeAppSensor) UniqueID() string {
	return "active_app_1"
}

func (s *activeAppSensor) UnitOfMeasurement() string {
	return ""
}

func (s *activeAppSensor) StateClass() string {
	return ""
}

func (s *activeAppSensor) EntityCategory() string {
	return ""
}

func (s *activeAppSensor) Disabled() bool {
	return s.disabled
}

func (s *activeAppSensor) Registered() bool {
	return s.registered
}

func (agent *Agent) runActiveAppSensor() {
	var encryptRequests = false
	if agent.config.secret != "" {
		encryptRequests = true
	}
	sensor := &activeAppSensor{
		state:  "Unknown",
		dataCh: make(chan interface{}),
	}
	sensorRequest := &sensorRequest{
		data:      sensor,
		encrypted: encryptRequests,
	}
	go device.ActiveAppUpdater(sensor.dataCh)
	for data := range sensor.dataCh {
		appName := data.(activeApp).Name()
		log.Debug().Caller().
			Msgf("Current active app %s.", appName)
		sensor.state = appName
		spew.Dump(sensorRequest)
		response := agent.updateActiveApp(sensorRequest)
		if response["success"].(bool) && !sensor.Registered() {
			sensor.registered = true
		}
	}
}

func (agent *Agent) updateActiveApp(request hass.Request) map[string]interface{} {
	agent.requestsCh <- request
	response := <-agent.responsesCh
	return response.(map[string]interface{})
	// switch v := response.(type) {
	// case error:
	// 	log.Error().Msg("Unable to update active app.")
	// 	return v
	// default:
	// 	log.Debug().Caller().Msg("Active app Updated.")
	// 	return v
	// }
}
