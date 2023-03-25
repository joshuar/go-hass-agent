package agent

import (
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
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

// func (l *location) MarshalJSON() ([]byte, error) {
// 	json, err := json.Marshal(&struct {
// 		Gps              []float64 `json:"gps"`
// 		GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
// 		Battery          int       `json:"battery,omitempty"`
// 		Speed            int       `json:"speed,omitempty"`
// 		Altitude         int       `json:"altitude,omitempty"`
// 		Course           int       `json:"course,omitempty"`
// 		VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
// 	}{
// 		Gps:              l.data.Gps(),
// 		GpsAccuracy:      l.data.GpsAccuracy(),
// 		Battery:          l.data.Battery(),
// 		Speed:            l.data.Speed(),
// 		Altitude:         l.data.Altitude(),
// 		Course:           l.data.Course(),
// 		VerticalAccuracy: l.data.VerticalAccuracy(),
// 	})
// 	if err != nil {
// 		log.Error().Msgf("Unable to convert location data to JSON: %v", err)
// 		return nil, err
// 	} else {
// 		return json, nil
// 	}
// }

func (a *Agent) runLocationWorker() {

	locationInfoCh := make(chan interface{})

	go device.LocationUpdater(a.App.UniqueID(), locationInfoCh)

	for loc := range locationInfoCh {
		log.Debug().Caller().
			Msgf("Got location %v.", loc.(locationData).Gps())
		l := &location{
			data: loc.(locationData),
		}
		a.updateLocation(l)
	}
}

func (a *Agent) updateLocation(r hass.Request) error {
	a.requestsCh <- r
	res := <-a.responsesCh
	switch v := res.(type) {
	case error:
		log.Error().Msg("Unable to update location.")
		return v
	default:
		log.Debug().Caller().Msg("Location Updated.")
		return nil
	}
}
