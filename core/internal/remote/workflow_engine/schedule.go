package workflow_engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/open-iga/core/internal/domain"
)

func (e *Engine) SetSchedule(ctx context.Context, workflowName string, schedule domain.WorkflowSchedule) error {
	writeCtx, cancel := dbos.WithTimeout(e.dbosCtx, requestTimeout)
	defer cancel()

	spec := dbos.ScheduleSpec{
		ScheduleName:      workflowName,
		Schedule:          schedule.Cron,
		WorkflowName:      scheduledWorkflowName,
		Workflow:          nil,
		WorkflowClassName: "",
		Context:           workflowName,
		AutomaticBackfill: false,
		CronTimezone:      scheduleTimezone,
		QueueName:         "",
		ApplicationName:   "",
	}
	err := dbos.ApplySchedules(writeCtx, []dbos.ScheduleSpec{spec})
	if errors.Is(err, dbos.ErrInvalidOption) {
		return fmt.Errorf("%w: schedule of %q: %w", domain.ErrInvalidWorkflow, workflowName, err)
	}

	if err != nil {
		return fmt.Errorf("save schedule of %q: %w", workflowName, err)
	}

	// Applying a schedule keeps the status it had, so the status is set on its own.
	setStatus := dbos.PauseSchedule
	if schedule.Enabled {
		setStatus = dbos.ResumeSchedule
	}

	if err := setStatus(writeCtx, workflowName); err != nil {
		return fmt.Errorf("set status of schedule %q: %w", workflowName, err)
	}

	return nil
}

func (e *Engine) GetSchedule(_ context.Context, workflowName string) (domain.WorkflowSchedule, error) {
	readCtx, cancel := dbos.WithTimeout(e.dbosCtx, requestTimeout)
	defer cancel()

	saved, err := dbos.GetSchedule(readCtx, workflowName)
	if errors.Is(err, dbos.ErrScheduleNotFound) {
		return domain.WorkflowSchedule{}, fmt.Errorf("%w: %s", domain.ErrScheduleNotFound, workflowName)
	}

	if err != nil {
		return domain.WorkflowSchedule{}, fmt.Errorf("get schedule of %q: %w", workflowName, err)
	}

	return domain.WorkflowSchedule{Cron: saved.Schedule, Enabled: saved.Status == dbos.ScheduleStatusActive}, nil
}
