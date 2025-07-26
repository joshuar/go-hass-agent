// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/id"
)

var (
	ErrRunFailed      = errors.New("failed to run scheduler")
	ErrScheduleFailed = errors.New("failed to schedule job")
)

type ManagerProps struct {
	scheduler quartz.Scheduler
}

var Manager *ManagerProps

func Start(ctx context.Context) error {
	scheduler, err := quartz.NewStdScheduler(
		quartz.WithOutdatedThreshold(50 * time.Second),
	)
	if err != nil {
		return errors.Join(ErrRunFailed, err)
	}

	Manager = &ManagerProps{
		scheduler: scheduler,
	}

	slogctx.FromCtx(ctx).Debug("Starting scheduler.")
	scheduler.Start(ctx)

	go func() {
		<-ctx.Done()
		slogctx.FromCtx(ctx).Debug("Stopping scheduler.")
		scheduler.Stop()
	}()

	return nil
}

func (m *ManagerProps) ScheduleJob(idPrefix id.Prefix, job quartz.Job, trigger quartz.Trigger) error {
	// Generate a job key.
	jobKey, err := id.NewID(idPrefix)
	if err != nil {
		return errors.Join(ErrScheduleFailed, err)
	}
	// Generate the job details.
	jobDetail := quartz.NewJobDetail(job, quartz.NewJobKey(jobKey))
	// Schedule the job.
	if err := m.scheduler.ScheduleJob(jobDetail, trigger); err != nil {
		return errors.Join(ErrScheduleFailed, err)
	}

	slog.Debug("Scheduled worker.",
		slog.String("job_key", jobKey),
		slog.String("job_desc", job.Description()))

	return nil
}

// PollTriggerWithJitter implements the quartz.Trigger interface; uses a fixed
// interval with an amount of jitter.
type PollTriggerWithJitter struct {
	Interval time.Duration
	Jitter   time.Duration
}

// Verify PollTriggerWithJitter satisfies the Trigger interface.
var _ quartz.Trigger = (*PollTriggerWithJitter)(nil)

// NewPollTriggerWithJitter returns a new PollTriggerWithJitter using the given interval.
func NewPollTriggerWithJitter(interval, jitter time.Duration) *PollTriggerWithJitter {
	return &PollTriggerWithJitter{
		Interval: interval,
		Jitter:   jitter,
	}
}

// NextFireTime returns the next time at which the PollTriggerWithJitter is scheduled to fire.
func (pt *PollTriggerWithJitter) NextFireTime(prev int64) (int64, error) {
	jitter := rand.NormFloat64()*float64(pt.Jitter) + float64(pt.Interval) // #nosec: G404
	next := prev + int64(jitter)

	return next, nil
}

// Description returns the description of the PollTriggerWithJitter.
func (pt *PollTriggerWithJitter) Description() string {
	return fmt.Sprintf("PollTriggerWithJitter%s%s", quartz.Sep, pt.Interval)
}
