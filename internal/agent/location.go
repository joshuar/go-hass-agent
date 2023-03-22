package agent

import (
	"context"
	"runtime"

	"github.com/carlmjohnson/requests"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/rs/zerolog/log"
)

type LocationInfo interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

func (a *Agent) runLocationUpdater() {

	locationInfoCh := make(chan interface{})

	switch os := runtime.GOOS; os {
	case "linux":
		go linux.LinuxLocationUpdater(a.App.UniqueID(), locationInfoCh)
	default:
		log.Error().Msg("Unsupported Operating System.")
	}

	for loc := range locationInfoCh {
		log.Debug().Caller().
			Msgf("Got location %v", loc.(LocationInfo).Gps())
		a.updateLocation(loc.(LocationInfo))
	}
}

func (a *Agent) updateLocation(l LocationInfo) error {
	req := &hass.GenericRequest{
		Type: hass.RequestTypeUpdateLocation,
		Data: struct {
			Gps              []float64 `json:"gps"`
			GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
			Battery          int       `json:"battery,omitempty"`
			Speed            int       `json:"speed,omitempty"`
			Altitude         int       `json:"altitude,omitempty"`
			Course           int       `json:"course,omitempty"`
			VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
		}{
			Gps:         l.Gps(),
			GpsAccuracy: l.GpsAccuracy(),
		},
	}
	res := &hass.GenericResponse{}
	ctx := context.Background()
	err := requests.
		URL(a.config.RestAPIURL).
		BodyJSON(&req).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return err
	} else {
		log.Debug().Msg("Location updated successfully")
		return nil
	}
}
