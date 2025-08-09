// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go tool golang.org/x/tools/cmd/stringer -type=hsiResult,hsiLevel -output hsi_generated.go -linecomment
package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/godbus/dbus/v5"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	fwupdmgrWorkerID   = "fwupdmgr_worker"
	fwupdmgrWorkerDesc = "fwupdmgr details"

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

var _ workers.EntityWorker = (*fwupdWorker)(nil)

var ErrInitFWUpdWorker = errors.New("could not init fwupdmgr worker")

type fwupdWorker struct {
	hostSecurityAttrs *dbusx.Data[[]map[string]dbus.Variant]
	hostSecurityID    *dbusx.Property[string]
	prefs             *preferences.CommonWorkerPrefs
	OutCh             chan models.Entity
	*models.WorkerMetadata
}

func (w *fwupdWorker) Execute(ctx context.Context) error {
	props, err := w.hostSecurityAttrs.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve security properties from fwupd: %w", err)
	}

	hsi, err := w.hostSecurityID.Get()
	if err != nil {
		return fmt.Errorf("could not retrieve security id from fwupd: %w", err)
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

	w.OutCh <- sensor.NewSensor(ctx,
		sensor.WithName("Firmware Security"),
		sensor.WithID("firmware_security"),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:security"),
		sensor.WithState(hsiID[0]),
		sensor.WithAttributes(attributes),
	)

	return nil
}

func (w *fwupdWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *fwupdWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *fwupdWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *fwupdWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	go func() {
		defer close(w.OutCh)
		if err := w.Execute(ctx); err != nil {
			slogctx.FromCtx(ctx).Warn("Failed to send fwupdmgr details",
				slog.Any("error", err))
		}
	}()
	return w.OutCh, nil
}

func NewfwupdWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitFWUpdWorker, linux.ErrNoSystemBus)
	}

	worker := &fwupdWorker{
		WorkerMetadata: models.SetWorkerMetadata(fwupdmgrWorkerID, fwupdmgrWorkerDesc),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitFWUpdWorker, err)
	}
	worker.prefs = prefs

	worker.hostSecurityAttrs = dbusx.NewData[[]map[string]dbus.Variant](bus,
		fwupdInterface, "/", fwupdInterface+"."+hostSecurityAttrsMethod)
	worker.hostSecurityID = dbusx.NewProperty[string](bus,
		"/", fwupdInterface, fwupdInterface+"."+hostSecurityIDProp)

	return worker, nil
}
