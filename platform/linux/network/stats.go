// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package network

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jsimonetti/rtnetlink"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

//go:generate go tool stringer -type=netStatsType -output stats.gen.go -linecomment
const (
	statBytesTx netStatsType = iota // Bytes Sent
	statBytesRx                     // Bytes Received
	statBytesTxRate                     // Bytes Sent Throughput
	statBytesRxRate                     // Bytes Received Throughput

)

const (
	statsWorkerID     = "network_stats"
	statsWorkerDesc   = "Network stats"
	statsWorkerPrefID = prefPrefix + "usage"

	countUnit = "B"
	rateUnit  = "B/s"

	rateInterval = 5 * time.Second
	rateJitter   = time.Second

	netRatesWorkerID = "network_stats_worker"

	totalsName = "Total"
)

var (
	_ quartz.Job                  = (*netStatsWorker)(nil)
	_ workers.PollingEntityWorker = (*netStatsWorker)(nil)
)

// StatsWorkerPrefs are the preferences for the stats worker.
type StatsWorkerPrefs struct {
	CommonPreferences

	UpdateInterval string `toml:"update_interval"`
}

// netStatsWorker is the object used for tracking network stats sensors. It
// holds a netlink connection and a map of links with their stats sensors.
type netStatsWorker struct {
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata

	statsSensors map[string]map[netStatsType]*netRate
	nlconn       *rtnetlink.Conn
	prefs        *StatsWorkerPrefs
	mu           sync.Mutex
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
		return nil, fmt.Errorf("unable to load net stats worker preferences: %w", err)
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

// Execute will gather all link stats and pass them through a channel on which the agent is listening for sensor updates.
func (w *netStatsWorker) Execute(ctx context.Context) error {
	delta := w.GetDelta()

	// Get all links.
	links, err := w.nlconn.Link.List()
	if err != nil {
		return fmt.Errorf("get links from netlink: %w", err)
	}

	// Get link stats, filtering links as per preferences.
	stats := w.getLinkStats(links)

	w.mu.Lock()
	defer w.mu.Unlock()

	// Counters for totals.
	var totalBytesRx, totalBytesTx uint64

	// For each link, update the link sensors with the new stats.
	for link := range slices.Values(stats) {
		var rate uint64

		linkName := link.name
		linkStats := link.stats
		totalBytesRx += linkStats.RXBytes
		totalBytesTx += linkStats.TXBytes

		// Generate bytesRx sensor entity for link.
		w.OutCh <- newStatsTotalEntity(ctx, linkName, statBytesRx, link.stats.RXBytes, getRXAttributes(linkStats))
		// Generate bytesTx sensor entity for link.
		w.OutCh <- newStatsTotalEntity(ctx, linkName, statBytesTx, link.stats.TXBytes, getTXAttributes(linkStats))
		// Create new trackers for the link's rates sensor entities if needed.
		if _, ok := w.statsSensors[linkName]; !ok {
			w.statsSensors[linkName] = newStatsRates()
		}
		// Generate bytesRxRate sensor entity for link.
		rate = w.statsSensors[linkName][statBytesRxRate].Calculate(linkStats.RXBytes, delta)
		w.OutCh <- newStatsRateEntity(ctx, linkName, statBytesRxRate, rate)
		// Generate bytesTxRate sensor entity for link.
		rate = w.statsSensors[linkName][statBytesTxRate].Calculate(linkStats.TXBytes, delta)
		w.OutCh <- newStatsRateEntity(ctx, linkName, statBytesTxRate, rate)
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
	// Generate total bytesRx sensor entity.
	w.OutCh <- newStatsTotalEntity(ctx, totalsName, statBytesRx, totalBytesRx, nil)
	// Generate total bytesTx sensor entity.
	w.OutCh <- newStatsTotalEntity(ctx, totalsName, statBytesTx, totalBytesTx, nil)
	// Generate total bytesRxRate sensor entity.
	rate = w.statsSensors[totalsName][statBytesRxRate].Calculate(totalStats.RXBytes, delta)
	w.OutCh <- newStatsRateEntity(ctx, totalsName, statBytesRxRate, rate)
	// Generate total bytesTxRate sensor entity.
	rate = w.statsSensors[totalsName][statBytesTxRate].Calculate(totalStats.TXBytes, delta)
	w.OutCh <- newStatsRateEntity(ctx, totalsName, statBytesTxRate, rate)

	return nil
}

func (w *netStatsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *netStatsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		return nil, fmt.Errorf("connect to netlink: %w", err)
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
		return w.OutCh, fmt.Errorf("schedule net stats worker: %w", err)
	}
	return w.OutCh, nil
}

// getLinkStats collates the network stats for all links. It filters
// links to those with stats, not in an exclusion list and currently active.
func (w *netStatsWorker) getLinkStats(links []rtnetlink.LinkMessage) []linkStats {
	allLinkStats := make([]linkStats, 0, len(links))

	for _, msg := range links {
		if msg.Attributes == nil {
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

// linkStats represents a link and its stats.
type linkStats struct {
	stats *rtnetlink.LinkStats64
	name  string
}

// netStatsType is the type of stat being tracked by an entity.
type netStatsType int

// Icon returns an material design icon representation of the network stat.
func (t netStatsType) Icon() string {
	switch t {
	case statBytesTx:
		return "mdi:upload-network"
	case statBytesRx:
		return "mdi:download-network"
	case statBytesTxRate:
		return "mdi:transfer-up"
	case statBytesRxRate:
		return "mdi:transfer-down"
	}

	return ""
}

// newStatsTotalEntity creates an entity for tracking total stats for a network
// device.
func newStatsTotalEntity(
	ctx context.Context,
	name string,
	entityType netStatsType,
	value uint64,
	attributes models.Attributes,
) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName(name+" "+entityType.String()),
		sensor.WithID(strings.ToLower(name)+"_"+strcase.ToSnake(entityType.String())),
		sensor.WithDeviceClass(class.SensorClassDataSize),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits(countUnit),
		sensor.WithIcon(entityType.Icon()),
		sensor.WithState(value),
		sensor.WithAttributes(attributes),
		sensor.WithDataSourceAttribute(linux.DataSrcNetlink),
		sensor.AsDiagnostic(),
	)
}

// newStatsTotalEntity creates an entity for tracking rate stats for a network device.
func newStatsRateEntity(ctx context.Context, name string, entityType netStatsType, value uint64) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName(name+" "+entityType.String()),
		sensor.WithID(strings.ToLower(name)+"_"+strcase.ToSnake(entityType.String())),
		sensor.WithDeviceClass(class.SensorClassDataRate),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits(rateUnit),
		sensor.WithIcon(entityType.Icon()),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcNetlink),
		sensor.AsDiagnostic(),
	)
}

type netRate struct {
	linux.RateValue[uint64]

	rateType netStatsType
}

func newStatsRates() map[netStatsType]*netRate {
	return map[netStatsType]*netRate{
		statBytesRxRate: {rateType: statBytesRxRate},
		statBytesTxRate: {rateType: statBytesTxRate},
	}
}

// getRXAttributes returns all sundry receive stats which can be added to a
// sensor as extra attributes.
func getRXAttributes(stats *rtnetlink.LinkStats64) models.Attributes {
	return models.Attributes{
		"receive_packets":       stats.RXPackets,
		"receive_errors":        stats.RXErrors,
		"receive_dropped":       stats.RXDropped,
		"multicast":             stats.Multicast,
		"collisions":            stats.Collisions,
		"receive_length_errors": stats.RXLengthErrors,
		"receive_over_errors":   stats.RXOverErrors,
		"receive_crc_errors":    stats.RXCRCErrors,
		"receive_frame_errors":  stats.RXFrameErrors,
		"receive_fifo_errors":   stats.RXFIFOErrors,
		"receive_missed_errors": stats.RXMissedErrors,
		"receive_compressed":    stats.RXCompressed,
		"transmit_compressed":   stats.TXCompressed,
		"receive_nohandler":     stats.RXNoHandler,
	}
}

// getTXAttributes returns all sundry transmit stats which can be added to a
// sensor as extra attributes.
func getTXAttributes(stats *rtnetlink.LinkStats64) models.Attributes {
	return models.Attributes{
		"transmit_packets":          stats.TXPackets,
		"transmit_errors":           stats.TXErrors,
		"transmit_dropped":          stats.TXDropped,
		"multicast":                 stats.Multicast,
		"collisions":                stats.Collisions,
		"transmit_aborted_errors":   stats.TXAbortedErrors,
		"transmit_carrier_errors":   stats.TXCarrierErrors,
		"transmit_fifo_errors":      stats.TXFIFOErrors,
		"transmit_heartbeat_errors": stats.TXHeartbeatErrors,
		"transmit_window_errors":    stats.TXWindowErrors,
		"transmit_compressed":       stats.TXCompressed,
	}
}
