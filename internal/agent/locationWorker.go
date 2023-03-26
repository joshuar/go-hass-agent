package agent

import (
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
)

type locationData interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

type location struct {
	data    locationData
	encrypt bool
}

func (l *location) RequestType() hass.RequestType {
	return hass.RequestTypeUpdateLocation
}

func (l *location) RequestData() interface{} {
	return struct {
		Gps              []float64 `json:"gps"`
		GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
		Battery          int       `json:"battery,omitempty"`
		Speed            int       `json:"speed,omitempty"`
		Altitude         int       `json:"altitude,omitempty"`
		Course           int       `json:"course,omitempty"`
		VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
	}{
		Gps:              l.data.Gps(),
		GpsAccuracy:      l.data.GpsAccuracy(),
		Battery:          l.data.Battery(),
		Speed:            l.data.Speed(),
		Altitude:         l.data.Altitude(),
		Course:           l.data.Course(),
		VerticalAccuracy: l.data.VerticalAccuracy(),
	}
}

func (l *location) IsEncrypted() bool {
	return l.encrypt
}
func (a *Agent) runLocationWorker(conn *hass.Conn) {

	locationInfoCh := make(chan interface{})

	go device.LocationUpdater(a.App.UniqueID(), locationInfoCh)

	for loc := range locationInfoCh {
		l := &location{
			data: loc.(locationData),
		}
		conn.SendRequest(l)
	}
}
