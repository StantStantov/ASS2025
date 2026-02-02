package dispatchers

import (
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/models"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type DispatchSystem struct {
	AlertsBuffer *buffer.BufferSystem

	Logger *logging.Logger
}

func NewDispatchSystem(
	buffer *buffer.BufferSystem,
	logger *logging.Logger,
) *DispatchSystem {
	system := &DispatchSystem{}

	system.AlertsBuffer = buffer

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "dispatch_system")
	})

	return system
}

func SaveAlerts(system *DispatchSystem, jobs ...*models.Job) {
	buffer.EnqueueIntoBuffer(system.AlertsBuffer, jobs...)

	logging.GetThenSendInfo(
		system.Logger,
		"saved jobs",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(jobs))
			amounts := make([]int, len(jobs))
			for i, job := range jobs {
				ids[i] = job.Id
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amount", amounts...)
			return nil
		},
	)
}

func GetFreeJobs(system *DispatchSystem, amount uint64) []*models.Job {
	jobs, _ := buffer.GetMultipleFreeFromBuffer(system.AlertsBuffer, amount)

	logging.GetThenSendInfo(
		system.Logger,
		"dispatched jobs",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(jobs))
			amounts := make([]int, len(jobs))
			for i, job := range jobs {
				ids[i] = job.Id
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amount", amounts...)
			return nil
		},
	)

	return jobs
}

func PutBusyJobs(system *DispatchSystem, jobs ...*models.Job)  {
	buffer.PutMultipleBusyIntoBuffer(system.AlertsBuffer, jobs...)

	logging.GetThenSendInfo(
		system.Logger,
		"returned jobs",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(jobs))
			amounts := make([]int, len(jobs))
			for i, job := range jobs {
				ids[i] = job.Id
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amount", amounts...)
			return nil
		},
	)
}
