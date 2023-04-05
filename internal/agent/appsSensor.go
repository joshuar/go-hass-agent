package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

// appSensor is a specific type of sensorState for app sensors
type appSensor sensorState

func newAppSensor(sensorType string) *appSensor {
	switch sensorType {
	case "ActiveApp":
		return &appSensor{
			name:     "active app",
			entityID: "active_app",
		}
	case "RunningApps":
		return &appSensor{
			name:       "running apps",
			entityID:   "running_apps",
			stateClass: measurement,
		}
	default:
		return nil
	}
}

// Ensure appSensor satisfies the sensor interface so it can be
// treated as a sensor

func (s *appSensor) Attributes() interface{} {
	return s.attributes
}

func (s *appSensor) DeviceClass() string {
	if s.deviceClass != 0 {
		return s.deviceClass.String()
	} else {
		return ""
	}
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
	return s.entityID
}

func (s *appSensor) UnitOfMeasurement() string {
	return ""
}

func (s *appSensor) StateClass() string {
	if s.stateClass != 0 {
		return s.stateClass.String()
	} else {
		return ""
	}
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

// Ensure that appSensor satisfies the hass.Request interface
// so its data can be sent as a request to HA

func (a *appSensor) RequestType() hass.RequestType {
	if a.Registered() {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (a *appSensor) RequestData() interface{} {
	return hass.MarshallSensorData(a)
}

func (s *appSensor) ResponseHandler(rawResponse interface{}) {
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

func (agent *Agent) runAppSensorWorker(ctx context.Context) {
	updateCh := make(chan interface{})
	defer close(updateCh)

	sensors := make(map[string]*appSensor)

	go device.AppUpdater(ctx, updateCh)

	for {
		select {
		case data := <-updateCh:
			update := data.(sensorUpdate)
			sensorID := update.Device() + update.Type()
			if _, ok := sensors[sensorID]; !ok {
				sensors[sensorID] = newAppSensor(update.Type())
			}
			sensors[sensorID].state = update.Value()
			sensors[sensorID].attributes = update.ExtraValues()
			go hass.APIRequest(ctx, sensors[sensorID])
		case <-ctx.Done():
			log.Debug().Caller().Msgf("Cleaning up app sensor.")
			return
		}
	}
}
