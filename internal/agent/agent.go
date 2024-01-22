// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/scripts"
	"github.com/joshuar/go-hass-agent/internal/tracker"
)

// Agent holds the data and structure representing an instance of the agent.
// This includes the data structure for the UI elements and tray and some
// strings such as app name and version.
type Agent struct {
	ui      UI
	done    chan struct{}
	Options *Options
}

// Options holds options taken from the command-line that was used to
// invoke go-hass-agent that are relevant for agent functionality.
type Options struct {
	ID, Server, Token       string
	Headless, ForceRegister bool
}

func New(o *Options) *Agent {
	a := &Agent{
		done:    make(chan struct{}),
		Options: o,
	}
	if !a.Options.Headless {
		a.ui = fyneui.NewFyneUI(a)
	}
	setupLogging()
	return a
}

// Run is the "main loop" of the agent. It sets up the agent, loads the config
// then spawns a sensor tracker and the workers to gather sensor data and
// publish it to Home Assistant.
func (agent *Agent) Run(cmd string) {
	var wg sync.WaitGroup
	var ctx context.Context
	var cancelFunc context.CancelFunc
	var err error

	var cfg config.Config
	configPath := filepath.Join(xdg.ConfigHome, agent.AppID())
	if cfg, err = config.Load(configPath); err != nil {
		log.Fatal().Err(err).Msg("Could not load config.")
	}

	var trk *tracker.SensorTracker
	if trk, err = tracker.NewSensorTracker(agent.AppID()); err != nil {
		log.Fatal().Err(err).Msg("Could not start sensor tracker.")
	}

	// Pre-flight: check if agent is registered. If not, run registration flow.
	var regWait sync.WaitGroup
	regWait.Add(1)
	go func() {
		defer regWait.Done()
		agent.checkRegistration(trk, cfg)
	}()

	if cmd == "go-hass-agent" {
		go func() {
			regWait.Wait()

			ctx, cancelFunc = setupContext(cfg)
			go func() {
				<-agent.done
				log.Debug().Msg("Agent done.")
				cancelFunc()
			}()

			// Start worker funcs for sensors.
			wg.Add(1)
			go func() {
				defer wg.Done()
				startWorkers(ctx, trk)
			}()
			// Start any scripts.
			wg.Add(1)
			go func() {
				defer wg.Done()
				scriptPath := filepath.Join(xdg.ConfigHome, agent.AppID(), "scripts")
				runScripts(ctx, scriptPath, trk)
			}()
			// Listen for notifications from Home Assistant.
			if !agent.IsHeadless() {
				wg.Add(1)
				go func() {
					defer wg.Done()
					agent.runNotificationsWorker(ctx)
				}()
			}
		}()
	}

	agent.handleSignals()

	if !agent.IsHeadless() {
		agent.ui.DisplayTrayIcon(agent, cfg, trk)
		agent.ui.Run(agent.done)
	}
	wg.Wait()
}

func (agent *Agent) Register() {
	var wg sync.WaitGroup
	var err error

	var cfg config.Config
	configPath := filepath.Join(xdg.ConfigHome, agent.AppID())
	if cfg, err = config.Load(configPath); err != nil {
		log.Fatal().Err(err).Msg("Could not load config.")
	}

	var trk *tracker.SensorTracker
	if trk, err = tracker.NewSensorTracker(agent.AppID()); err != nil {
		log.Fatal().Err(err).Msg("Could not start sensor tracker.")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.checkRegistration(trk, cfg)
	}()

	go func() {
		wg.Wait()
		log.Debug().Msg("Registration complete.")
		agent.Stop()
	}()

	agent.ui.Run(agent.done)
}

// handleSignals will handle Ctrl-C of the agent.
func (agent *Agent) handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer close(agent.done)
		<-c
		log.Debug().Msg("Ctrl-C pressed.")
	}()
}

// IsHeadless returns a bool indicating whether the agent is running in
// "headless" mode (i.e., without a GUI) or not.
func (agent *Agent) IsHeadless() bool {
	return agent.Options.Headless
}

// AppID returns the "application ID". Currently, this ID is just used to
// indicate whether the agent is running in debug mode or not.
func (agent *Agent) AppID() string {
	return agent.Options.ID
}

// Stop will close the agent's done channel which indicates to any goroutines it
// is time to clean up and exit.
func (agent *Agent) Stop() {
	log.Debug().Msg("Stopping agent.")
	close(agent.done)
}

// startWorkers will call all the sensor worker functions that have been defined
// for this device.
func startWorkers(ctx context.Context, trk *tracker.SensorTracker) {
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
func runScripts(ctx context.Context, path string, trk *tracker.SensorTracker) {
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

// setupContext embeds the config object in a context which allows access to it
// from any functions that inherit this context. This is used early in the agent
// start up to ensure all subsequent functionality can access config details as
// needed.
func setupContext(cfg config.Config) (context.Context, context.CancelFunc) {
	baseCtx, cancelFunc := context.WithCancel(context.Background())
	agentCtx := config.EmbedInContext(baseCtx, cfg)
	return agentCtx, cancelFunc
}

// setupLogging will attempt to create and then write logging to a file. If it
// cannot do this, logging will only be available on stdout.
func setupLogging() {
	logFile := filepath.Join(xdg.StateHome, "go-hass-app.log")
	logWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Error().Err(err).
			Msg("Unable to open log file for writing.")
	} else {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		multiWriter := zerolog.MultiLevelWriter(consoleWriter, logWriter)
		log.Logger = log.Output(multiWriter)
	}
}
