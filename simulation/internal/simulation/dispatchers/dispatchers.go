package dispatchers

import (
	"StantStantov/ASS/internal/simulation/buffer"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"StantStantov/ASS/internal/simulation/pools"

	"github.com/StantStantov/rps/swamp/behaivors/buffers"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type DispatchSystem struct {
	AlertsBuffer *buffer.BufferSystem
	AlertsPool   *pools.PoolSystem

	Metrics *metrics.MetricsSystem
	Logger  *logging.Logger
}

func NewDispatchSystem(
	buffer *buffer.BufferSystem,
	pool *pools.PoolSystem,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *DispatchSystem {
	system := &DispatchSystem{}

	system.AlertsBuffer = buffer
	system.AlertsPool = pool

	system.Metrics = metrics
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
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amount", amounts...)

			return nil
		},
	)
}

func GetFreeJobs(system *DispatchSystem, setBuffer *buffers.SetBuffer[models.Job, uint64]) {
	logging.GetThenSendDebug(
		system.Logger,
		"going to dispatch jobs",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Integer(event, "jobs.requested_amount", len(setBuffer.Array))

			return nil
		},
	)

	ids := make([]uint64, cap(setBuffer.Array))
	idsBuffer := &buffers.SetBuffer[uint64, uint64]{Array: ids}
	pools.GetFromPool(system.AlertsPool, idsBuffer)
	buffer.GetMultipleFromBuffer(system.AlertsBuffer, setBuffer, ids...)

	logging.GetThenSendInfo(
		system.Logger,
		"dispatched jobs",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Integer(event, "jobs.requested_amount", len(setBuffer.Array))
			logfmt.Unsigneds(event, "jobs.ids", ids...)

			return nil
		},
	)
}

func PutBusyJobs(system *DispatchSystem, jobs ...models.Job) {
	ids := make([]uint64, len(jobs))
	ids = models.JobsToIds(jobs, ids)
	pools.RemoveFromPool(system.AlertsPool, ids...)

	logging.GetThenSendInfo(
		system.Logger,
		"returned jobs",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", ids...)

			return nil
		},
	)
}
