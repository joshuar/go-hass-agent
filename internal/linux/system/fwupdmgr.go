// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go run golang.org/x/tools/cmd/stringer -type=hsiResult,hsiLevel -output hsi_generated.go -linecomment
package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	fwupdmgrWorkerID = "system_info"

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

	hsiSensor := sensor.Entity{
		Name:     "Firmware Security",
		Category: types.CategoryDiagnostic,
		State: &sensor.State{
			ID:         "firmware_security",
			Value:      hsiID[0],
			Icon:       "mdi:security",
			Attributes: make(map[string]any),
		},
	}

	for _, prop := range props {
		var (
			summary string
			result  hsiResult
		)

		summary, _ = dbusx.VariantToValue[string](prop["Summary"])
		result, _ = dbusx.VariantToValue[hsiResult](prop["HsiResult"])

		hsiSensor.Attributes[summary] = result.String()
	}

	return []sensor.Entity{hsiSensor}, nil
}

func NewfwupdWorker(ctx context.Context) (*linux.OneShotSensorWorker, error) {
	worker := linux.NewOneShotWorker(fwupdmgrWorkerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	worker.OneShotType = &fwupdWorker{
		hostSecurityAttrs: dbusx.NewData[[]map[string]dbus.Variant](bus, fwupdInterface, "/", fwupdInterface+"."+hostSecurityAttrsMethod),
		hostSecurityID:    dbusx.NewProperty[string](bus, "/", fwupdInterface, fwupdInterface+"."+hostSecurityIDProp),
	}

	return worker, nil
}
