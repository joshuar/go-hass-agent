// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go run golang.org/x/tools/cmd/stringer -type=hsiResult,hsiLevel -output hsi_generated.go -linecomment
package system

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	fwupdmgrWorkerID = "fwupdmgr_worker"

	fwupdInterface          = "org.freedesktop.fwupd"
	hostSecurityAttrsMethod = "GetHostSecurityAttrs"
	hostSecurityIDProp      = "HostSecurityId"
)

const (
	ResultUnknown      hsiResult = iota // Not Known
	ResultEnabled                       // Enabled
	ResultNotEnabled                    // Not Enabled
	ResultValid                         // Valid
	ResultNotValid                      // Not Valid
	ResultLocked                        // Locked
	ResultNotLocked                     // Not Locked
	ResultEncrypted                     // Encrypted
	ResultNotEncrypted                  // Not Encrypted
	ResultTainted                       // Tainted
	ResultNotTainted                    // Not Tainted
	ResultFound                         // Found
	ResultNotFound                      // Not Found
	ResultSupported                     // Supported
	ResultNotSupported                  // Not Supported

)

type hsiResult uint32

const (
	hsi0 hsiLevel = iota // HSI:0 (Insecure State)
	hsi1                 // HSI:1 (Critical State)
	hsi2                 // HSI:2 (Risky State)
	hsi3                 // HSI:3 (Protected State)
	hsi4                 // HSI:4 (Secure State)
	hsi5                 // HSI:5 (Secure Proven State)
)

type hsiLevel uint32

var ErrInitFWUpdWorker = errors.New("could not init fwupdmgr worker")

type fwupdWorker struct {
	hostSecurityAttrs *dbusx.Data[[]map[string]dbus.Variant]
	hostSecurityID    *dbusx.Property[string]
}

//nolint:errcheck
func (w *fwupdWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	props, err := w.hostSecurityAttrs.Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve security properties from fwupd: %w", err)
	}

	hsi, err := w.hostSecurityID.Get()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve security id from fwupd: %w", err)
	}

	hsiID := strings.Split(hsi, " ")
	attributes := make(map[string]any)

	for _, prop := range props {
		var (
			summary string
			result  hsiResult
		)

		if summaryRaw, found := prop["Summary"]; found {
			summary, _ = dbusx.VariantToValue[string](summaryRaw)
		}

		if resultRaw, found := prop["HsiResult"]; found {
			result, _ = dbusx.VariantToValue[hsiResult](resultRaw)
		}

		if summary != "" && result != ResultUnknown {
			attributes[summary] = result.String()
		}
	}

	return []sensor.Entity{
			sensor.NewSensor(
				sensor.WithName("Firmware Security"),
				sensor.WithID("firmware_security"),
				sensor.AsDiagnostic(),
				sensor.WithState(
					sensor.WithIcon("mdi:security"),
					sensor.WithValue(hsiID[0]),
					sensor.WithAttributes(attributes),
				),
			),
		},
		nil
}

func (w *fwupdWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *fwupdWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewfwupdWorker(ctx context.Context) (*linux.OneShotSensorWorker, error) {
	fwupdWorker := &fwupdWorker{}

	prefs, err := preferences.LoadWorker(fwupdWorker)
	if err != nil {
		return nil, errors.Join(ErrInitFWUpdWorker, err)
	}

	//nolint:nilnil
	if prefs.IsDisabled() {
		return nil, nil
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitFWUpdWorker, linux.ErrNoSystemBus)
	}

	fwupdWorker.hostSecurityAttrs = dbusx.NewData[[]map[string]dbus.Variant](bus,
		fwupdInterface, "/", fwupdInterface+"."+hostSecurityAttrsMethod)
	fwupdWorker.hostSecurityID = dbusx.NewProperty[string](bus,
		"/", fwupdInterface, fwupdInterface+"."+hostSecurityIDProp)

	worker := linux.NewOneShotSensorWorker(fwupdmgrWorkerID)
	worker.OneShotSensorType = fwupdWorker

	return worker, nil
}
