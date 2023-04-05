package agent

import (
	"context"
	"fmt"
	"math"

	"github.com/gobeam/stringy"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=batteryDeviceClass -output batterySensor_iota.go
const (
	battery batteryDeviceClass = iota
	temperature
	power
	state
)

// batteryDeviceClass is the Device Class of the sensor, derived from the state
// type.
type batteryDeviceClass int

// batterySensor is a specific type of sensorState for battery sensors
type batterySensor sensorState

type batterySensorUpdate interface {
	ID() string
	Type() device.BatteryProp
	Value() interface{}
	ExtraValues() interface{}
}

func newBatterySensor(batteryID string, sensorType device.BatteryProp) *batterySensor {
	var sensorName, sensorID, stateClass string
	var deviceClass batteryDeviceClass
	str := stringy.New(sensorType.String())
	switch sensorType {
	case device.Percentage:
		sensorName = batteryID + " Battery Level"
		sensorID = batteryID + "_battery_level"
		stateClass = "measurement"
		deviceClass = battery
	case device.Temperature:
		sensorName = batteryID + " Battery Temperature"
		sensorID = batteryID + "_battery_temperature"
		stateClass = "measurement"
		deviceClass = temperature
	case device.EnergyRate:
		sensorName = batteryID + " Battery Power"
		sensorID = batteryID + "_battery_power"
		stateClass = "measurement"
		deviceClass = power
	default:
		sensorName = batteryID + " Battery " + sensorType.String()
		sensorID = batteryID + "_" + str.SnakeCase().ToLower()
		stateClass = ""
		deviceClass = state
	}
	return &batterySensor{
		name:        sensorName,
		entityID:    sensorID,
		deviceClass: deviceClass,
		stateClass:  stateClass,
	}
}

// Ensure a batterySensor satisfies the sensor interface so it can be treated as
// a sensor

func (b *batterySensor) Attributes() interface{} {
	return b.attributes
}

func (b *batterySensor) DeviceClass() string {
	switch b.deviceClass {
	case battery, temperature, power:
		return b.deviceClass.(batteryDeviceClass).String()
	default:
		return ""
	}
}

func (b *batterySensor) Icon() string {
	switch b.deviceClass {
	case battery:
		if b.state.(float64) >= 95 {
			return "mdi:battery"
		} else {
			return fmt.Sprintf("mdi:battery-%d", int(math.Round(b.state.(float64)/10)*10))
		}
	case power:
		if math.Signbit(b.state.(float64)) {
			return "mdi:battery-minus"
		} else {
			return "mdi:battery-plus"
		}
	default:
		return "mdi:battery"
	}
}

func (b *batterySensor) Name() string {
	return b.name
}

func (b *batterySensor) State() interface{} {
	switch b.deviceClass {
	case battery:
		return b.state.(float64)
	default:
		return b.state
	}
}

func (b *batterySensor) SensorType() string {
	return "sensor"
}

func (b *batterySensor) UniqueID() string {
	return b.entityID
}

func (b *batterySensor) UnitOfMeasurement() string {
	switch b.deviceClass {
	case battery:
		return "%"
	case temperature:
		return "°C"
	case power:
		return "W"
	default:
		return ""
	}
}

func (b *batterySensor) StateClass() string {
	return b.stateClass
}

func (b *batterySensor) EntityCategory() string {
	return "diagnostic"
}

func (b *batterySensor) Disabled() bool {
	return b.disabled
}

func (b *batterySensor) Registered() bool {
	return b.registered
}

// Ensure that a batterySensor satisfies the hass.Request interface so its data
// can be sent as a request to HA

func (b *batterySensor) RequestType() hass.RequestType {
	if b.Registered() {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (b *batterySensor) RequestData() interface{} {
	return hass.MarshallSensorData(b)
}

func (b *batterySensor) ResponseHandler(rawResponse interface{}) {
	if rawResponse == nil {
		log.Debug().Caller().Msg("No response data.")
	} else {
		response := rawResponse.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !b.registered {
				b.registered = true
				log.Debug().Caller().Msgf("Sensor %s registered.", b.Name())
			}
		}
		if v, ok := response[b.entityID]; ok {
			status := v.(map[string]interface{})
			if !status["success"].(bool) {
				error := status["error"].(map[string]interface{})
				log.Error().Msgf("Could not update sensor %s, %s: %s", b.Name(), error["code"], error["message"])
			} else {
				log.Debug().Msgf("Sensor %s updated. State is now: %v", b.Name(), b.State())
			}
			if v, ok := status["is_disabled"]; ok {
				switch v.(bool) {
				case true:
					log.Debug().Msgf("Sensor %s has been disabled.", b.Name())
					b.disabled = true
				case false:
					log.Debug().Msgf("Sensor %s has been enabled.", b.Name())
					b.disabled = false
				}
			}
		}
	}
}

func (agent *Agent) runBatterySensorWorker(ctx context.Context) {

	updateCh := make(chan interface{})
	defer close(updateCh)

	sensors := make(map[string]*batterySensor)

	go device.BatteryUpdater(ctx, updateCh)

	for {
		select {
		case data := <-updateCh:
			update := data.(batterySensorUpdate)
			sensorID := update.ID() + update.Type().String()
			if _, ok := sensors[sensorID]; !ok {
				sensors[sensorID] = newBatterySensor(update.ID(), update.Type())
			}
			sensors[sensorID].state = update.Value()
			sensors[sensorID].attributes = update.ExtraValues()
			go hass.APIRequest(ctx, sensors[sensorID])
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Cleaning up battery sensors.")
			return
		}
	}
}
