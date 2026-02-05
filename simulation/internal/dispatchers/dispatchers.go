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

func SaveAlerts(system *DispatchSystem, jobs ...models.Job) {
	buffer.AddIntoBuffer(system.AlertsBuffer, jobs...)
	pools.MoveIfNewIntoPool(system.AlertsPool, jobs...)

	logging.GetThenSendInfo(
		system.Logger,
		"saved jobs",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(jobs))
			ids = models.JobsToIds(jobs, ids)
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

func GetFreeJobs(system *DispatchSystem, setBuffer []models.Job) []models.Job {
	ids := make([]uint64, len(setBuffer))
	ids = pools.GetFromPool(system.AlertsPool, ids)
	setBuffer = buffer.GetMultipleFromBuffer(system.AlertsBuffer, setBuffer, ids...)

	logging.GetThenSendInfo(
		system.Logger,
		"dispatched jobs",
		func(event *logging.Event, level logging.Level) error {
			amounts := make([]int, len(setBuffer))
			for i, job := range setBuffer {
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amount", amounts...)

			return nil
		},
	)

	return setBuffer
}

func PutBusyJobs(system *DispatchSystem, jobs ...models.Job) {
	ids := make([]uint64, len(jobs))
	ids = models.JobsToIds(jobs, ids)
	pools.RemoveFromPool(system.AlertsPool, ids...)

	logging.GetThenSendInfo(
		system.Logger,
		"returned jobs",
		func(event *logging.Event, level logging.Level) error {
			amounts := make([]int, len(jobs))
			for i, job := range jobs {
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amount", amounts...)

			return nil
		},
	)
}
