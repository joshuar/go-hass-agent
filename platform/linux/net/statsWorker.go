// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package net

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jsimonetti/rtnetlink"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	statsWorkerID     = "network_stats"
	statsWorkerDesc   = "Network stats"
	statsWorkerPrefID = prefPrefix + "usage"
)

var (
	_ quartz.Job                  = (*netStatsWorker)(nil)
	_ workers.PollingEntityWorker = (*netStatsWorker)(nil)
)

var ErrInitStatsWorker = errors.New("could not init network stats worker")

type StatsWorkerPrefs struct {
	WorkerPrefs
	UpdateInterval string `toml:"update_interval"`
}

// netStatsWorker is the object used for tracking network stats sensors. It
// holds a netlink connection and a map of links with their stats sensors.
type netStatsWorker struct {
	statsSensors map[string]map[netStatsType]*netRate
	nlconn       *rtnetlink.Conn
	prefs        *StatsWorkerPrefs
	mu           sync.Mutex
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata
}

//nolint:funlen
func (w *netStatsWorker) Execute(ctx context.Context) error {
	delta := w.GetDelta()

	// Get all links.
	links, err := w.getLinks()
	if err != nil {
		return fmt.Errorf("error accessing netlink: %w", err)
	}

	// Get link stats, filtering "uninteresting" links.
	stats := w.getLinkStats(links)

	w.mu.Lock()
	defer w.mu.Unlock()

	// Counters for totals.
	var totalBytesRx, totalBytesTx uint64

	// For each link, update the link sensors with the new stats.
	for link := range slices.Values(stats) {
		var rate uint64

		name := link.name
		stats := link.stats
		totalBytesRx += stats.RXBytes
		totalBytesTx += stats.TXBytes

		// Skip ignored devices.
		if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
			return strings.HasPrefix(name, e)
		}) {
			continue
		}

		// Generate bytesRecv sensor entity for link.
		w.OutCh <- newStatsTotalEntity(ctx, name, bytesRecv, link.stats.RXBytes, getRXAttributes(stats))
		// Generate bytesSent sensor entity for link.
		w.OutCh <- newStatsTotalEntity(ctx, name, bytesSent, link.stats.TXBytes, getTXAttributes(stats))
		// Create new trackers for the link's rates sensor entities if needed.
		if _, ok := w.statsSensors[name]; !ok {
			w.statsSensors[name] = newStatsRates()
		}
		// Generate bytesRecvRate sensor entity for link.
		rate = w.statsSensors[name][bytesRecvRate].Calculate(stats.RXBytes, delta)
		w.OutCh <- newStatsRateEntity(ctx, name, bytesRecvRate, rate)
		// Generate bytesSentRate sensor entity for link.
		rate = w.statsSensors[name][bytesSentRate].Calculate(stats.TXBytes, delta)
		w.OutCh <- newStatsRateEntity(ctx, name, bytesSentRate, rate)
	}

	if len(stats) == 0 {
		return nil
	}

	var rate uint64
	// Create a pseudo total stats link stats object.
	totalStats := &rtnetlink.LinkStats64{
		RXBytes: totalBytesRx,
		TXBytes: totalBytesTx,
	}
	// Generate total bytesRecv sensor entity.
	w.OutCh <- newStatsTotalEntity(ctx, totalsName, bytesRecv, totalBytesRx, nil)
	// Generate total bytesSent sensor entity.
	w.OutCh <- newStatsTotalEntity(ctx, totalsName, bytesSent, totalBytesTx, nil)
	// Generate total bytesRecvRate sensor entity.
	rate = w.statsSensors[totalsName][bytesRecvRate].Calculate(totalStats.RXBytes, delta)
	w.OutCh <- newStatsRateEntity(ctx, totalsName, bytesRecvRate, rate)
	// Generate total bytesSentRate sensor entity.
	rate = w.statsSensors[totalsName][bytesSentRate].Calculate(totalStats.TXBytes, delta)
	w.OutCh <- newStatsRateEntity(ctx, totalsName, bytesSentRate, rate)

	return nil
}

func (w *netStatsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *netStatsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		return nil, errors.Join(ErrInitStatsWorker,
			fmt.Errorf("could not connect to netlink: %w", err))
	}

	w.nlconn = conn

	go func() {
		<-ctx.Done()

		if err = w.nlconn.Close(); err != nil {
			slogctx.FromCtx(ctx).Debug("Could not close netlink connection.",
				slog.String("worker", netRatesWorkerID),
				slog.Any("error", err))
		}
	}()

	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
}

// getLinks returns all available links on this device. If a problem occurred, a
// non-nil error is returned.
func (w *netStatsWorker) getLinks() ([]rtnetlink.LinkMessage, error) {
	links, err := w.nlconn.Link.List()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of devices from netlink: %w", err)
	}

	return links, nil
}

// getLinkStats collates the network stats for all links. It filters
// links to those with stats, not in an exclusion list and currently active.
func (w *netStatsWorker) getLinkStats(links []rtnetlink.LinkMessage) []linkStats {
	allLinkStats := make([]linkStats, 0, len(links))

	for _, msg := range links {
		if msg.Attributes == nil {
			continue
		}

		// Ignore loopback.
		if msg.Attributes.Name == loopbackDeviceName {
			continue
		}

		// Skip ignored devices.
		if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
			return strings.HasPrefix(msg.Attributes.Name, e)
		}) {
			continue
		}

		// Ignore devices that are not currently active.
		if *msg.Attributes.Carrier == 0 {
			continue
		}

		name := msg.Attributes.Name
		stats := msg.Attributes.Stats64

		if stats32 := msg.Attributes.Stats; stats == nil && stats32 != nil {
			stats = &rtnetlink.LinkStats64{
				RXPackets:          uint64(stats32.RXPackets),
				TXPackets:          uint64(stats32.TXPackets),
				RXBytes:            uint64(stats32.RXBytes),
				TXBytes:            uint64(stats32.TXBytes),
				RXErrors:           uint64(stats32.RXErrors),
				TXErrors:           uint64(stats32.TXErrors),
				RXDropped:          uint64(stats32.RXDropped),
				TXDropped:          uint64(stats32.TXDropped),
				Multicast:          uint64(stats32.Multicast),
				Collisions:         uint64(stats32.Collisions),
				RXLengthErrors:     uint64(stats32.RXLengthErrors),
				RXOverErrors:       uint64(stats32.RXOverErrors),
				RXCRCErrors:        uint64(stats32.RXCRCErrors),
				RXFrameErrors:      uint64(stats32.RXFrameErrors),
				RXFIFOErrors:       uint64(stats32.RXFIFOErrors),
				RXMissedErrors:     uint64(stats32.RXMissedErrors),
				TXAbortedErrors:    uint64(stats32.TXAbortedErrors),
				TXCarrierErrors:    uint64(stats32.TXCarrierErrors),
				TXFIFOErrors:       uint64(stats32.TXFIFOErrors),
				TXHeartbeatErrors:  uint64(stats32.TXHeartbeatErrors),
				TXWindowErrors:     uint64(stats32.TXWindowErrors),
				RXCompressed:       uint64(stats32.RXCompressed),
				TXCompressed:       uint64(stats32.TXCompressed),
				RXNoHandler:        uint64(stats32.RXNoHandler),
				RXOtherhostDropped: 0,
			}
		}

		if stats != nil {
			allLinkStats = append(allLinkStats, linkStats{
				name:  name,
				stats: stats,
			})
		}
	}

	return allLinkStats
}

// NewNetStatsWorker sets up a sensor worker that tracks network stats.
func NewNetStatsWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &netStatsWorker{
		WorkerMetadata:          models.SetWorkerMetadata(statsWorkerID, statsWorkerDesc),
		statsSensors:            make(map[string]map[netStatsType]*netRate),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}
	worker.statsSensors[totalsName] = newStatsRates()

	defaultPrefs := &StatsWorkerPrefs{
		UpdateInterval: rateInterval.String(),
	}
	defaultPrefs.IgnoredDevices = defaultIgnoredDevices
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(statsWorkerPrefID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitStatsWorker, err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", netRatesWorkerID),
			slog.String("given_interval", worker.prefs.UpdateInterval),
			slog.String("default_interval", rateInterval.String()))

		pollInterval = rateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, rateJitter)

	return worker, nil
}
