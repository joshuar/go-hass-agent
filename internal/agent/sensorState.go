package agent

import (
	"context"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type sensorState struct {
	deviceClass hass.SensorDeviceClass
	stateClass  hass.SensorStateClass
	sensorType  hass.SensorType
	state       interface{}
	stateUnits  string
	attributes  interface{}
	icon        string
	name        string
	entityID    string
	category    string
	disabled    bool
	registered  bool
}

// sensorState implements hass.Sensor

func (s *sensorState) Attributes() interface{} {
	return s.attributes
}

func (s *sensorState) DeviceClass() string {
	if s.deviceClass != 0 {
		return s.deviceClass.String()
	} else {
		return ""
	}
}

func (s *sensorState) Icon() string {
	return s.icon
}

func (s *sensorState) Name() string {
	return s.name
}

func (s *sensorState) State() interface{} {
	return s.state
}

func (s *sensorState) Type() string {
	switch s.sensorType {
	case hass.TypeSensor:
		return "sensor"
	case hass.TypeBinary:
		return "binary_sensor"
	default:
		log.Debug().Caller().Msgf("Invalid or unknown sensor type %v", s.sensorType)
		return ""
	}
}

func (s *sensorState) UniqueID() string {
	return s.entityID
}

func (s *sensorState) UnitOfMeasurement() string {
	return s.stateUnits
}

func (s *sensorState) StateClass() string {
	if s.stateClass != 0 {
		return s.stateClass.String()
	} else {
		return ""
	}
}

func (s *sensorState) EntityCategory() string {
	return s.category
}

func (s *sensorState) Disabled() bool {
	return s.disabled
}

func (s *sensorState) Registered() bool {
	return s.registered
}

// sensorState implements hass.Request

func (b *sensorState) RequestType() hass.RequestType {
	if b.registered {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (b *sensorState) RequestData() interface{} {
	return hass.MarshalSensorData(b)
}

func (b *sensorState) ResponseHandler(rawResponse interface{}) {
	if rawResponse == nil || len(rawResponse.(map[string]interface{})) == 0 {
		log.Debug().Caller().Msg("No response data.")
	} else {
		response := rawResponse.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !b.registered {
				b.registered = true
				log.Debug().Caller().Msgf("Sensor %s registered.", b.name)
			}
		}
		if v, ok := response[b.entityID]; ok {
			status := v.(map[string]interface{})
			if !status["success"].(bool) {
				error := status["error"].(map[string]interface{})
				log.Error().Msgf("Could not update sensor %s, %s: %s", b.name, error["code"], error["message"])
			} else {
				log.Debug().Msgf("Sensor %s updated. State is now: %v", b.name, b.state)
			}
			if v, ok := status["is_disabled"]; ok {
				switch v.(bool) {
				case true:
					log.Debug().Msgf("Sensor %s has been disabled.", b.name)
					b.disabled = true
				case false:
					log.Debug().Msgf("Sensor %s has been enabled.", b.name)
					b.disabled = false
				}
			}
		}
	}
}

func newSensor(s hass.SensorUpdate) *sensorState {
	sensor := &sensorState{
		deviceClass: s.DeviceClass(),
		stateClass:  s.StateClass(),
		sensorType:  s.SensorType(),
		state:       s.State(),
		attributes:  s.Attributes(),
		icon:        s.Icon(),
		stateUnits:  s.Units(),
		category:    s.Category(),
		registered:  false,
		disabled:    false,
	}
	if s.Device() != "" {
		sensor.name = s.Device() + " " + strcase.ToDelimited(s.Name(), ' ')
		sensor.entityID = s.Device() + "_" + strings.ToLower(strcase.ToSnake(s.Name()))
	} else {
		sensor.name = strcase.ToDelimited(s.Name(), ' ')
		sensor.entityID = strings.ToLower(strcase.ToSnake(s.Name()))
	}
	return sensor
}

func (s *sensorState) updateSensor(ctx context.Context, update hass.SensorUpdate) {
	s.state = update.State()
	s.attributes = update.Attributes()
	s.icon = update.Icon()
	go hass.APIRequest(ctx, s)
}

func TrackSensors(ctx context.Context) {

	updateCh := make(chan interface{})
	// defer close(updateCh)
	doneCh := make(chan struct{})

	sensors := make(map[string]*sensorState)

	go device.AppUpdater(ctx, updateCh, doneCh)
	go device.BatteryUpdater(ctx, updateCh, doneCh)

	for {
		select {
		case data := <-updateCh:
			sensorID := data.(hass.SensorUpdate).Device() + data.(hass.SensorUpdate).Name()
			if _, ok := sensors[sensorID]; !ok {
				sensors[sensorID] = newSensor(data.(hass.SensorUpdate))
				log.Debug().Caller().Msgf("New sensor discovered: %s", sensors[sensorID].name)
				// log.Debug().Msgf("Attributes %v", sensors[sensorID].Attributes())
				// log.Debug().Msgf("DeviceClass %s", sensors[sensorID].DeviceClass())
				// log.Debug().Msgf("Icon %s", sensors[sensorID].Icon())
				// log.Debug().Msgf("Name %s", sensors[sensorID].Name())
				// log.Debug().Msgf("State %v", sensors[sensorID].State())
				// log.Debug().Msgf("Type %s", sensors[sensorID].Type())
				// log.Debug().Msgf("UniqueID %s", sensors[sensorID].UniqueID())
				// log.Debug().Msgf("Unit %s", sensors[sensorID].UnitOfMeasurement())
				// log.Debug().Msgf("StateClass %s", sensors[sensorID].StateClass())
				// log.Debug().Msgf("Category %s", sensors[sensorID].EntityCategory())
				// log.Debug().Msgf("Disabled? %v", sensors[sensorID].Disabled())
				// log.Debug().Msgf("Registered? %v", sensors[sensorID].Registered())
				go hass.APIRequest(ctx, sensors[sensorID])
			} else {
				sensors[sensorID].updateSensor(ctx, data.(hass.SensorUpdate))
			}
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Stopping sensor tracking.")
			close(doneCh)
			return
		}
	}
}
