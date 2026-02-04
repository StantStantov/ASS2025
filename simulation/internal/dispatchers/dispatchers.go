package dispatchers

import (
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type DispatchSystem struct {
	AlertsBuffer *buffer.BufferSystem
	AlertsPool   *pools.PoolSystem

	Logger *logging.Logger
}

func NewDispatchSystem(
	buffer *buffer.BufferSystem,
	pool *pools.PoolSystem,
	logger *logging.Logger,
) *DispatchSystem {
	system := &DispatchSystem{}

	system.AlertsBuffer = buffer
	system.AlertsPool = pool

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "dispatch_system")
	})

	return system
}

func SaveAlerts(system *DispatchSystem, jobs ...*models.Job) {
	buffer.AddIntoBuffer(system.AlertsBuffer, jobs...)
	pools.MoveIfNewIntoPool(system.AlertsPool, jobs...)

	logging.GetThenSendInfo(
		system.Logger,
		"saved jobs",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(jobs))
			ids = models.JobsPtrToIds(jobs, ids)
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

func GetFreeJobs(system *DispatchSystem, maxAmount uint64) []*models.Job {
	jobs := make([]*models.Job, maxAmount)
	jobs = pools.GetFromPool(system.AlertsPool, jobs)

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

func PutBusyJobs(system *DispatchSystem, jobs ...*models.Job) {
	pools.RemoveFromPool(system.AlertsPool, jobs...)

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
