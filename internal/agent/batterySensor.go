package agent

import (
	"context"
	"fmt"
	"math"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=batteryDeviceClass,batteryType -output batterySensor_string.go
const (
	battery batteryDeviceClass = iota
	temperature
	power
	state

	unknown batteryType = iota
	linePower
	generic
	ups
	monitor
	mouse
	keyboard
	pda
	phone
)

type batteryDeviceClass int
type batteryType int

// BatteryState is an interface that represents the state
// of a device battery at a particular point in time
type BatteryState interface {
	LevelPercent() float64
	Temperature() float64
	Health() string
	Power() float64
	Voltage() float64
	Energy() float64
	ChargerType() string
	State() string
	ID() string
	Type() interface{}
}

type batterySensor sensorState

func newBatterySensor(batteryID string, sensorType batteryDeviceClass) *batterySensor {
	var sensorName, sensorID, stateClass string
	switch sensorType {
	case battery:
		sensorName = batteryID + " Battery Level"
		sensorID = batteryID + "_battery_level"
		stateClass = "measurement"
	case temperature:
		sensorName = batteryID + " Battery Temperature"
		sensorID = batteryID + "_battery_temperature"
		stateClass = "measurement"
	case power:
		sensorName = batteryID + " Battery Power"
		sensorID = batteryID + "_battery_power"
		stateClass = "measurement"
	case state:
		sensorName = batteryID + " Battery State"
		sensorID = batteryID + "_battery_state"
		stateClass = ""
	default:
		return nil
	}
	return &batterySensor{
		name:        sensorName,
		entityID:    sensorID,
		deviceClass: sensorType,
		stateClass:  stateClass,
	}

}

// Ensure a batterySensor satisfies the sensor interface so it can be
// treated as a sensor

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

func (b *batterySensor) HandleAPIResponse(rawResponse interface{}) {
	if rawResponse == nil {
		log.Debug().Caller().Msg("No response data.")
	} else {
		response := rawResponse.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !b.Registered() {
				b.registered = true
				log.Debug().Caller().Msgf("Sensor %s registered.", b.Name())
			}
		}
		if v, ok := response[b.UniqueID()]; ok {
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

// Ensure that a batterySensor satisfies the hass.Request interface
// so its data can be sent as a request to HA

func (b *batterySensor) RequestType() hass.RequestType {
	if b.Registered() {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (b *batterySensor) RequestData() interface{} {
	return hass.MarshallSensorData(b)
}

// batteryTracker keeps track of a particular battery
// it handles creating and updating the individual sensors
// associated with the battery in HA

type batteryTracker struct {
	updateCh    chan BatteryState
	currentInfo BatteryState
	sensors     map[batteryDeviceClass]*batterySensor
}

func newBatteryTracker(ctx context.Context, batteryID string) *batteryTracker {
	newTracker := &batteryTracker{
		updateCh: make(chan BatteryState),
		sensors:  make(map[batteryDeviceClass]*batterySensor),
	}

	newTracker.sensors[battery] = newBatterySensor(batteryID, battery)
	newTracker.sensors[temperature] = newBatterySensor(batteryID, temperature)
	newTracker.sensors[power] = newBatterySensor(batteryID, power)
	newTracker.sensors[state] = newBatterySensor(batteryID, state)
	go newTracker.monitor(ctx)
	return newTracker
}

func (tracker *batteryTracker) monitor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-tracker.updateCh:
			tracker.currentInfo = update
			tracker.sensors[battery].state = tracker.currentInfo.LevelPercent()
			tracker.sensors[temperature].state = tracker.currentInfo.Temperature()
			tracker.sensors[power].state = tracker.currentInfo.Power()
			tracker.sensors[state].state = tracker.currentInfo.State()
			tracker.sensors[power].attributes = struct {
				Voltage float64 `json:"voltage"`
				Energy  float64 `json:"energy"`
			}{
				Voltage: tracker.currentInfo.Voltage(),
				Energy:  tracker.currentInfo.Energy(),
			}
		}
	}
}

func (tracker *batteryTracker) UpdateHass(ctx context.Context) {
	go hass.APIRequest(ctx, tracker.sensors[power], tracker.sensors[power].HandleAPIResponse)
	go hass.APIRequest(ctx, tracker.sensors[temperature], tracker.sensors[temperature].HandleAPIResponse)
	go hass.APIRequest(ctx, tracker.sensors[battery], tracker.sensors[battery].HandleAPIResponse)
	go hass.APIRequest(ctx, tracker.sensors[state], tracker.sensors[state].HandleAPIResponse)
}

func (agent *Agent) runBatterySensorWorker(ctx context.Context) {

	updateCh := make(chan interface{})
	defer close(updateCh)

	batteries := make(map[string]*batteryTracker)

	go device.BatteryUpdater(ctx, updateCh)

	for {
		select {
		case i := <-updateCh:
			info := i.(BatteryState)
			if _, ok := batteries[info.ID()]; ok {
				batteries[info.ID()].updateCh <- info
			} else {
				batteries[info.ID()] = newBatteryTracker(ctx, info.ID())
				batteries[info.ID()].updateCh <- info
			}
			batteries[info.ID()].UpdateHass(ctx)
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Cleaning up battery sensors.")
			return
		}
	}
}
