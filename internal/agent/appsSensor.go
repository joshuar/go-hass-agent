package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type activeApp interface {
	Name() string
	Attributes() interface{}
}

type runningApps interface {
	Count() int
	Attributes() interface{}
}

type appSensor struct {
	name       string
	id         string
	state      interface{}
	stateClass string
	attributes interface{}
	disabled   bool
	registered bool
}

func (s *appSensor) Attributes() interface{} {
	return s.attributes
}

func (s *appSensor) DeviceClass() string {
	return ""
}

func (s *appSensor) Icon() string {
	return "mdi:application"
}

func (s *appSensor) Name() string {
	return s.name
}

func (s *appSensor) State() interface{} {
	return s.state
}

func (s *appSensor) SensorType() string {
	return "sensor"
}

func (s *appSensor) UniqueID() string {
	return s.id
}

func (s *appSensor) UnitOfMeasurement() string {
	return ""
}

func (s *appSensor) StateClass() string {
	return s.stateClass
}

func (s *appSensor) EntityCategory() string {
	return ""
}

func (s *appSensor) Disabled() bool {
	return s.disabled
}

func (s *appSensor) Registered() bool {
	return s.registered
}

func (s *appSensor) HandleAPIResponse(rawResponse interface{}) {
	if rawResponse == nil {
		log.Debug().Caller().Msg("No response data.")
	} else {
		response := rawResponse.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !s.Registered() {
				s.registered = true
				log.Debug().Caller().Msgf("Sensor %s registered.", s.Name())
			}
		}
		if v, ok := response[s.UniqueID()]; ok {
			status := v.(map[string]interface{})
			if !status["success"].(bool) {
				error := status["error"].(map[string]interface{})
				log.Error().Msgf("Could not update sensor %s, %s: %s", s.Name(), error["code"], error["message"])
			} else {
				log.Debug().Msgf("Sensor %s updated. State is now: %v", s.Name(), s.State())
			}
			if v, ok := status["is_disabled"]; ok {
				switch v.(bool) {
				case true:
					log.Debug().Msgf("Sensor %s has been disabled.", s.Name())
					s.disabled = true
				case false:
					log.Debug().Msgf("Sensor %s has been enabled.", s.Name())
					s.disabled = false
				}
			}
		}
	}
}

func (agent *Agent) runAppSensorWorker() {
	var encryptRequests = false
	if agent.config.secret != "" {
		encryptRequests = true
	}

	updateCh := make(chan interface{})
	defer close(updateCh)

	ctx := context.Background()

	deviceName, _ := agent.GetDeviceDetails()

	activeAppSensor := &appSensor{
		state:      "Unknown",
		name:       deviceName + " Active App",
		id:         deviceName + "_active_app",
		registered: false,
		disabled:   false,
	}

	runningAppsSensor := &appSensor{
		state:      "Unknown",
		name:       deviceName + " Running Apps",
		id:         deviceName + "_running_apps",
		stateClass: "measurement",
		registered: false,
		disabled:   false,
	}

	go device.AppUpdater(updateCh)

	for data := range updateCh {
		activeAppSensor.state = data.(activeApp).Name()
		activeAppSensor.attributes = data.(activeApp).Attributes()
		go hass.APIRequest(ctx, agent.config.APIURL, &sensorRequest{
			data:      activeAppSensor,
			encrypted: encryptRequests,
		}, activeAppSensor.HandleAPIResponse)

		runningAppsSensor.state = data.(runningApps).Count()
		runningAppsSensor.attributes = data.(runningApps).Attributes()
		go hass.APIRequest(ctx, agent.config.APIURL, &sensorRequest{
			data:      runningAppsSensor,
			encrypted: encryptRequests,
		}, runningAppsSensor.HandleAPIResponse)
	}
}
