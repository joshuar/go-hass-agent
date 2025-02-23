// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/internal/components/id"
	"github.com/joshuar/go-hass-agent/internal/components/logging"
)

var (
	ErrRunFailed      = errors.New("failed to run scheduler")
	ErrScheduleFailed = errors.New("failed to schedule job")
)

type ManagerProps struct {
	scheduler quartz.Scheduler
	logger    *slog.Logger
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
		logger:    logging.FromContext(ctx).WithGroup("scheduler"),
	}

	Manager.logger.Debug("Starting scheduler.")
	scheduler.Start(ctx)

	go func() {
		<-ctx.Done()
		Manager.logger.Debug("Stopping scheduler.")
		scheduler.Stop()
	}()

	return nil
}

func (m *ManagerProps) ScheduleJob(job quartz.Job, trigger quartz.Trigger) error {
	// Generate a job key.
	jobKey, err := id.NewID(id.SchedulerJob)
	if err != nil {
		return errors.Join(ErrScheduleFailed, err)
	}
	// Generate the job details.
	jobDetail := quartz.NewJobDetail(job, quartz.NewJobKey(jobKey))
	// Schedule the job.
	if err := m.scheduler.ScheduleJob(jobDetail, trigger); err != nil {
		return errors.Join(ErrScheduleFailed, err)
	}

	m.logger.Debug("Scheduled job.",
		slog.String("job_key", jobKey),
		slog.String("job_desc", job.Description()))

	return nil
}
