package linux

import (
	"github.com/maltegrosse/go-geoclue2"
	"github.com/rs/zerolog/log"
)

type linuxLocation struct {
	latitude  float64
	longitude float64
	accuracy  float64
	speed     float64
}

func (l *linuxLocation) Gps() []float64 {
	return []float64{l.latitude, l.longitude}
}

func (l *linuxLocation) GpsAccuracy() int {
	return int(l.accuracy)
}

func (l *linuxLocation) Battery() int {
	return 0
}

func (l *linuxLocation) Speed() int {
	return int(l.speed)
}

func (l *linuxLocation) Altitude() int {
	return 0
}

func (l *linuxLocation) Course() int {
	return 0
}

func (l *linuxLocation) VerticalAccuracy() int {
	return 0
}

func LinuxLocationUpdater(locationInfoCh chan interface{}) {

	locationInfo := &linuxLocation{}

	gcm, err := geoclue2.NewGeoclueManager()
	if err != nil {
		log.Error().Msg(err.Error())
	}
	client, err := gcm.GetClient()
	if err != nil {
		log.Error().Msg(err.Error())
	}
	err = client.SetDesktopId("foo")
	err = client.Start()
	if err != nil {
		log.Error().Msg(err.Error())
	}

	c := client.SubscribeLocationUpdated()
	for v := range c {
		_, location, err := client.ParseLocationUpdated(v)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		locationInfo.latitude, err = location.GetLatitude()
		if err != nil {
			log.Error().Msg(err.Error())
		}
		locationInfo.longitude, err = location.GetLongitude()
		if err != nil {
			log.Error().Msg(err.Error())
		}
		locationInfo.accuracy, err = location.GetAccuracy()
		if err != nil {
			log.Error().Msg(err.Error())
		}

		locationInfo.speed, err = location.GetSpeed()
		if err != nil {
			log.Error().Msg(err.Error())
		}

		locationInfoCh <- locationInfo
	}

}
