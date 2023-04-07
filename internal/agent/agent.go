package agent

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/jeandeaual/go-locale"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	Name      = "go-hass-agent"
	Version   = "0.0.1"
	fyneAppID = "com.github.joshuar.go-hass-agent"
)

type Agent struct {
	App           fyne.App
	Tray          fyne.Window
	Name, Version string
	MsgPrinter    *message.Printer
}

func Run() {
	ctx, cancelfunc := context.WithCancel(context.Background())
	deviceCtx := device.SetupContext(ctx)
	start(deviceCtx)
	cancelfunc()
}

func start(ctx context.Context) {
	agent := &Agent{
		App:     newUI(),
		Name:    Name,
		Version: Version,
	}

	userLocales, err := locale.GetLocales()
	if err != nil {
		log.Warn().Msg("Could not find a suitable locale. Defaulting to English.")
		agent.MsgPrinter = message.NewPrinter(message.MatchLanguage(language.English.String()))
	} else {
		agent.MsgPrinter = message.NewPrinter(message.MatchLanguage(userLocales...))
		log.Debug().Caller().Msgf("Setting language to %v.", userLocales)
	}
	agent.setupSystemTray()

	var once sync.Once

	appConfig := &config.AppConfig{}
	once.Do(func() { appConfig = agent.loadConfig(ctx) })
	workerCtx := config.NewContext(ctx, appConfig)

	go agent.runNotificationsWorker(workerCtx)
	// go agent.runLocationWorker(workerCtx)
	go agent.trackSensors(workerCtx)

	// var wg sync.WaitGroup

	// wg.Add(1)
	// func(ctx context.Context) {
	// 	defer wg.Done()
	// 	agent.runNotificationsWorker(ctx)
	// }(workerCtx)

	// wg.Add(1)
	// func(ctx context.Context) {
	// 	defer wg.Done()
	// 	agent.runLocationWorker(ctx)
	// }(workerCtx)

	agent.App.Run()
	// wg.Wait()
	agent.stop()
}

func (agent *Agent) stop() {
	log.Debug().Caller().Msg("Shutting down agent.")
	agent.Tray.Close()
}

// TrackSensors should be run in a goroutine and is responsible for creating,
// tracking and update HA with all sensors provided from the platform/device.
func (agent *Agent) trackSensors(ctx context.Context) {

	deviceAPI, deviceAPIExists := device.FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not retrieve deviceAPI from context.")
		return
	}

	updateCh := make(chan interface{})
	// defer close(updateCh)
	doneCh := make(chan struct{})

	sensors := make(map[string]*sensorState)

	go func() {
		for {
			select {
			case data := <-updateCh:
				switch data := data.(type) {
				case hass.SensorUpdate:
					sensorID := data.Device() + data.Name()
					if _, ok := sensors[sensorID]; !ok {
						sensors[sensorID] = newSensor(data)
						log.Debug().Caller().Msgf("New sensor discovered: %s", sensors[sensorID].name)
						go hass.APIRequest(ctx, sensors[sensorID])
					} else {
						sensors[sensorID].updateSensor(ctx, data)
					}
				case hass.LocationUpdate:
					l := &location{
						data: data,
					}
					go hass.APIRequest(ctx, l)
				}
			case <-ctx.Done():
				log.Debug().Caller().
					Msg("Stopping sensor tracking.")
				close(doneCh)
				return
			}
		}
	}()

	// go device.LocationUpdater(agent.App.UniqueID(), updateCh)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		device.LocationUpdater(agent.App.UniqueID(), updateCh, doneCh)
	}()

	for name, workerFunction := range deviceAPI.SensorInfo.Get() {
		wg.Add(1)
		log.Debug().Caller().
			Msgf("Running a worker for %s.", name)
		go func(worker func(context.Context, chan interface{}, chan struct{})) {
			defer wg.Done()
			worker(ctx, updateCh, doneCh)
		}(workerFunction)
	}
	wg.Wait()
}
