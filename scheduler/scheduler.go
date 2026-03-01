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

	"github.com/joshuar/go-hass-agent/id"
)

var (
	ErrRunFailed      = errors.New("failed to run scheduler")
	ErrScheduleFailed = errors.New("failed to schedule job")
)

type manager struct {
	quartz.Scheduler
}

var mgr manager

// Start wil start the scheduler component of the agent.
func Start(ctx context.Context) error {
	misfiredCh := make(chan quartz.ScheduledJob)
	scheduler, err := quartz.NewStdScheduler(
		quartz.WithOutdatedThreshold(50*time.Second),
		// quartz.WithLogger(&logger{Logger: slogctx.FromCtx(ctx)}),
		quartz.WithMisfiredChan(misfiredCh),
	)
	if err != nil {
		return errors.Join(ErrRunFailed, err)
	}

	mgr = manager{
		Scheduler: scheduler,
	}

	// Run goroutine to log misfired jobs.
	go func() {
		for misfiredJob := range misfiredCh {
			slogctx.FromCtx(ctx).Debug("Job misfired.",
				slog.String("job_id", misfiredJob.JobDetail().JobKey().String()),
				slog.String("job_description", misfiredJob.JobDetail().Job().Description()),
			)
		}
	}()

	scheduler.Start(ctx)
	slogctx.FromCtx(ctx).Info("Scheduler started.",
		slog.Time("timestamp", time.Now()),
	)

	go func() {
		<-ctx.Done()
		scheduler.Stop()
		slogctx.FromCtx(ctx).Info("Scheduler stopped.",
			slog.Time("timestamp", time.Now()),
		)
	}()

	return nil
}

func ScheduleJob(idPrefix id.Prefix, job quartz.Job, trigger quartz.Trigger) error {
	// Generate a job key.
	jobKey, err := id.NewID(idPrefix)
	if err != nil {
		return errors.Join(ErrScheduleFailed, err)
	}
	// Generate the job details.
	jobDetail := quartz.NewJobDetail(job, quartz.NewJobKey(jobKey))
	// Schedule the job.
	err = mgr.ScheduleJob(jobDetail, trigger)
	if err != nil {
		return errors.Join(ErrScheduleFailed, err)
	}
	return nil
}

func IsStarted() bool {
	return mgr.IsStarted()
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

// type logger struct {
// 	*slog.Logger
// }

// func (l *logger) Trace(msg string, args ...any) {
// 	l.Debug(msg, args...)
// }
