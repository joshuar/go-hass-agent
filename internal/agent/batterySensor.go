package agent

import (
	"context"
	"fmt"
	"math"

	"github.com/davecgh/go-spew/spew"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=batteryDeviceClass
const (
	battery batteryDeviceClass = iota
	temperature
	power
	none
)

type batteryDeviceClass int

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
}

// batterySensor implements the sensor interface to
// create a battery sensor in HA based on its deviceClass
type batterySensor struct {
	deviceClass     batteryDeviceClass
	state           interface{}
	attributes      interface{}
	name            string
	entityID        string
	disabled        bool
	registered      bool
	encryptRequests bool
}

func newBatterySensor(batteryID string, sensorType batteryDeviceClass, encryptRequests bool) *batterySensor {
	var sensorName, sensorID string
	switch sensorType {
	case battery:
		sensorName = batteryID + " Battery Level"
		sensorID = batteryID + "_battery_level"
	case temperature:
		sensorName = batteryID + " Battery Temperature"
		sensorID = batteryID + "_battery_temperature"
	case power:
		sensorName = batteryID + " Battery Power"
		sensorID = batteryID + "_battery_power"
	default:
		return nil
	}
	return &batterySensor{
		name:            sensorName,
		entityID:        sensorID,
		encryptRequests: encryptRequests,
		deviceClass:     sensorType,
	}

}

// Ensure a batterySensor satisfies the sensor interface so it can be
// treated as a sensor

func (b *batterySensor) Attributes() interface{} {
	return b.attributes
}

func (b *batterySensor) DeviceClass() string {
	return b.deviceClass.String()
}

func (b *batterySensor) Icon() string {
	switch b.deviceClass {
	case battery:
		return fmt.Sprintf("mdi:battery-%d", int(math.Round(b.state.(float64)/0.1)*10))
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
		return int(b.state.(float64) * 100)
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
	return "measurement"
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
	return MarshallSensorData(b)
}

func (b *batterySensor) IsEncrypted() bool {
	return b.encryptRequests
}

// batteryTracker keeps track of a particular battery
// it handles creating and updating the individual sensors
// associated with the battery in HA

type batteryTracker struct {
	updateCh    chan BatteryState
	currentInfo BatteryState
	sensors     map[batteryDeviceClass]*batterySensor
}

func newBatteryTracker(batteryID string, encryptRequests bool) *batteryTracker {
	newTracker := &batteryTracker{
		updateCh: make(chan BatteryState),
		sensors:  make(map[batteryDeviceClass]*batterySensor),
	}
	newTracker.sensors[battery] = newBatterySensor(batteryID, battery, encryptRequests)
	newTracker.sensors[temperature] = newBatterySensor(batteryID, temperature, encryptRequests)
	newTracker.sensors[power] = newBatterySensor(batteryID, power, encryptRequests)
	go newTracker.monitor()
	return newTracker
}

func (tracker *batteryTracker) monitor() {
	for update := range tracker.updateCh {
		tracker.currentInfo = update
		tracker.sensors[battery].state = tracker.currentInfo.LevelPercent()
		tracker.sensors[temperature].state = tracker.currentInfo.Temperature()
		tracker.sensors[power].state = tracker.currentInfo.Power()
		tracker.sensors[power].attributes = struct {
			Voltage float64 `json:"voltage"`
			Energy  float64 `json:"energy"`
		}{
			Voltage: tracker.currentInfo.Voltage(),
			Energy:  tracker.currentInfo.Energy(),
		}
		spew.Dump(tracker.sensors[power].RequestData())
	}
}

func (tracker *batteryTracker) UpdateHass(ctx context.Context, url string) {
	go hass.APIRequest(ctx, url, tracker.sensors[power], tracker.sensors[power].HandleAPIResponse)
	go hass.APIRequest(ctx, url, tracker.sensors[temperature], tracker.sensors[temperature].HandleAPIResponse)
	go hass.APIRequest(ctx, url, tracker.sensors[battery], tracker.sensors[battery].HandleAPIResponse)
}

func (agent *Agent) runBatterySensorWorker() {
	var encryptRequests = false
	if agent.config.secret != "" {
		encryptRequests = true
	}

	updateCh := make(chan interface{})
	defer close(updateCh)

	// deviceName, _ := agent.GetDeviceDetails()
	apiURL := agent.config.APIURL

	batteries := make(map[string]*batteryTracker)

	ctx := context.Background()
	go device.BatteryUpdater(updateCh)

	for i := range updateCh {
		info := i.(BatteryState)
		spew.Dump(info)
		if _, ok := batteries[info.ID()]; ok {
			batteries[info.ID()].updateCh <- info
		} else {
			batteries[info.ID()] = newBatteryTracker(info.ID(), encryptRequests)
			batteries[info.ID()].updateCh <- info
		}
		batteries[info.ID()].UpdateHass(ctx, apiURL)
	}
}
