// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package user

import (
	"context"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	loginBasePath        = "/org/freedesktop/login1"
	loginBaseInterface   = "org.freedesktop.login1"
	managerInterface     = loginBaseInterface + ".Manager"
	sessionAddedSignal   = "SessionNew"
	sessionRemovedSignal = "SessionRemoved"
	listSessionsMethod   = managerInterface + ".ListSessions"
)

type usersSensor struct {
	userNames []string
	linux.Sensor
}

func (s *usersSensor) Attributes() any {
	return struct {
		DataSource string   `json:"Data Source"`
		Usernames  []string `json:"Usernames"`
	}{
		DataSource: linux.DataSrcDbus,
		Usernames:  s.userNames,
	}
}

func (s *usersSensor) updateUsers(ctx context.Context) {
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(loginBasePath).
		Destination(loginBaseInterface)
	userData, err := dbusx.GetData[[][]any](req, listSessionsMethod)
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve users from D-Bus.")
		return
	}
	s.Value = len(userData)
	var users []string
	for _, u := range userData {
		if user, ok := u[2].(string); ok {
			users = append(users, user)
		}
	}
	s.userNames = users
}

func newUsersSensor() *usersSensor {
	s := &usersSensor{}
	s.SensorTypeValue = linux.SensorUsers
	s.UnitsString = "users"
	s.IconString = "mdi:account"
	s.StateClassValue = types.StateClassMeasurement
	return s
}

func Updater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	u := newUsersSensor()

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{sessionAddedSignal, sessionRemovedSignal},
		Interface: managerInterface,
		Path:      loginBasePath,
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create active users D-Bus watch.")
		close(sensorCh)
		return sensorCh
	}

	go func() {
		defer close(sensorCh)
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped users sensors.")
				return
			case event := <-events:
				spew.Dump(event)
				if !strings.Contains(event.Signal, sessionAddedSignal) && !strings.Contains(event.Signal, sessionRemovedSignal) {
					continue
				}
				go func() {
					u.updateUsers(ctx)
					sensorCh <- u
				}()
			}
		}
	}()
	// Send an initial sensor update.
	go func() {
		u.updateUsers(ctx)
		sensorCh <- u
	}()

	return sensorCh
}
