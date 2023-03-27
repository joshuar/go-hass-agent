package agent

import (
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

func (s *appSensor) handleResponse(rawResponse interface{}) {
	if rawResponse == nil {
		log.Debug().Msg("No response data.")
		return
	}
	response := rawResponse.(map[string]interface{})
	if v, ok := response["success"]; ok {
		if v.(bool) && !s.Registered() {
			s.registered = true
			log.Debug().Caller().Msgf("Sensor %s registered.", s.name)
		}
	}
	if v, ok := response[s.UniqueID()]; ok {
		status := v.(map[string]interface{})
		if !status["success"].(bool) {
			error := status["error"].(map[string]interface{})
			log.Error().Msgf("Could not update sensor %s, %s: %s", s.name, error["code"], error["message"])
		} else {
			log.Debug().Msgf("Sensor %s updated. State is now: %v", s.name, s.state)
		}
		if v, ok := status["is_disabled"]; ok {
			switch v.(bool) {
			case true:
				log.Debug().Msgf("Sensor %s has been disabled.", s.name)
				s.disabled = true
			case false:
				log.Debug().Msgf("Sensor %s has been enabled.", s.name)
				s.disabled = false
			}
		}
	}
}

func (agent *Agent) runActiveAppSensor(conn *hass.Conn) {
	var encryptRequests = false
	if agent.config.secret != "" {
		encryptRequests = true
	}

	updateCh := make(chan interface{})
	defer close(updateCh)

	activeAppSensor := &appSensor{
		state:      "Unknown",
		name:       "Active App",
		id:         "active_app_2",
		registered: false,
		disabled:   false,
	}

	runningAppsSensor := &appSensor{
		state:      "Unknown",
		name:       "Running Apps",
		id:         "running_apps",
		stateClass: "measurement",
		registered: false,
		disabled:   false,
	}

	go device.AppUpdater(updateCh)

	for data := range updateCh {
		// switch sensor := data.(type) {
		// case activeApp:
		var response interface{}

		activeAppSensor.state = data.(activeApp).Name()
		activeAppSensor.attributes = data.(activeApp).Attributes()
		response = conn.SendRequest(&sensorRequest{
			data:      activeAppSensor,
			encrypted: encryptRequests,
		})
		activeAppSensor.handleResponse(response)
		// case runningApps:
		runningAppsSensor.state = data.(runningApps).Count()
		runningAppsSensor.attributes = data.(runningApps).Attributes()
		response = conn.SendRequest(&sensorRequest{
			data:      runningAppsSensor,
			encrypted: encryptRequests,
		})
		runningAppsSensor.handleResponse(response)
		// }
	}
}
