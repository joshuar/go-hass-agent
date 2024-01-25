// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/scripts"
	"github.com/joshuar/go-hass-agent/internal/tracker"
)

// runWorkers will call all the sensor worker functions that have been defined
// for this device.
func runWorkers(ctx context.Context, trk SensorTracker) {
	workerFuncs := sensorWorkers()
	workerFuncs = append(workerFuncs, device.ExternalIPUpdater)
	d := newDevice(ctx)
	workerCtx := d.Setup(ctx)

	var wg sync.WaitGroup
	var outCh []<-chan tracker.Sensor

	log.Debug().Msg("Starting worker funcs.")
	for i := 0; i < len(workerFuncs); i++ {
		outCh = append(outCh, workerFuncs[i](workerCtx))
	}

	wg.Add(1)
	go func() {
		log.Debug().Msg("Listening for sensor updates.")
		defer wg.Done()
		for s := range tracker.MergeSensorCh(ctx, outCh...) {
			go func(s tracker.Sensor) {
				trk.UpdateSensors(ctx, s)
			}(s)
		}
	}()
	wg.Add(1)
	go func() {
		log.Debug().Msg("Listening for location updates.")
		defer wg.Done()
		for l := range locationWorker()(workerCtx) {
			go func(l *hass.LocationData) {
				trk.UpdateSensors(ctx, l)
			}(l)
		}
	}()

	wg.Wait()
}

// runScripts will retrieve all scripts that the agent can run and queue them up
// to be run on their defined schedule using the cron scheduler. It also sets up
// a channel to receive script output and send appropriate sensor objects to the
// tracker.
func runScripts(ctx context.Context, path string, trk SensorTracker) {
	allScripts, err := scripts.FindScripts(path)
	switch {
	case err != nil:
		log.Error().Err(err).Msg("Error getting scripts.")
		return
	case len(allScripts) == 0:
		log.Debug().Msg("Could not find any script files.")
		return
	}
	c := cron.New()
	var outCh []<-chan tracker.Sensor
	for _, s := range allScripts {
		schedule := s.Schedule()
		if schedule != "" {
			_, err := c.AddJob(schedule, s)
			if err != nil {
				log.Warn().Err(err).Str("script", s.Path()).
					Msg("Unable to schedule script.")
				break
			}
			outCh = append(outCh, s.Output)
			log.Debug().Str("schedule", schedule).Str("script", s.Path()).
				Msg("Added script sensor.")
		}
	}
	log.Debug().Msg("Starting cron scheduler for script sensors.")
	c.Start()
	go func() {
		for s := range tracker.MergeSensorCh(ctx, outCh...) {
			go func(s tracker.Sensor) {
				trk.UpdateSensors(ctx, s)
			}(s)
		}
	}()
	<-ctx.Done()
	log.Debug().Msg("Stopping cron scheduler for script sensors.")
	cronCtx := c.Stop()
	<-cronCtx.Done()
}

func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	log.Debug().Msg("Listening for notifications.")

	notifyCh := make(chan [2]string)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopping notification handler.")
				return
			case n := <-notifyCh:
				agent.ui.DisplayNotification(n[0], n[1])
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopping websocket.")
				return
			default:
				api.StartWebsocket(ctx, notifyCh)
			}
		}
	}()

	wg.Wait()
}
