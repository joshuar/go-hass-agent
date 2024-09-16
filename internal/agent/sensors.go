// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver,unexported-return
package agent

import (
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type version string

func (v *version) Name() string { return "Go Hass Agent Version" }

func (v *version) ID() string { return "agent_version" }

func (v *version) Icon() string { return "mdi:face-agent" }

func (v *version) SensorType() types.SensorClass { return types.Sensor }

func (v *version) DeviceClass() types.DeviceClass { return 0 }

func (v *version) StateClass() types.StateClass { return 0 }

func (v *version) State() any { return preferences.AppVersion }

func (v *version) Units() string { return "" }

func (v *version) Category() string { return types.CategoryDiagnostic }

func (v *version) Attributes() map[string]any { return nil }
