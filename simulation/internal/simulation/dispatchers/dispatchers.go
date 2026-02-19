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

func SaveAlerts(system *DispatchSystem, ids []models.AgentId, alertsBatches [][]models.MachineInfo) {
	logging.GetThenSendDebug(
		system.Logger,
		"going to save jobs",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Integer(event, "jobs.requested_amount", len(ids))

			return nil
		},
	)

	buffer.AddIntoBuffer(system.AlertsBuffer, ids, alertsBatches)
	pools.MoveIfNewIntoPool(system.AlertsPool, ids)

	logging.GetThenSendInfo(
		system.Logger,
		"saved jobs",
		func(event *logging.Event, level logging.Level) error {
			amounts := make([]int, len(alertsBatches))
			for i, alerts := range alertsBatches {
				amounts[i] = len(alerts)
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
	alertsBatches := make([][]models.MachineInfo, cap(setBuffer.Array))
	alertsBuffer := &buffers.SetBuffer[[]models.MachineInfo, uint64]{Array: alertsBatches}
	pools.GetFromPool(system.AlertsPool, idsBuffer)
	buffer.GetMultipleFromBuffer(system.AlertsBuffer, alertsBuffer, ids...)

	minLength := min(idsBuffer.Length, alertsBuffer.Length)
	for i := range minLength {
		job := models.Job{
			Id:     ids[i],
			Alerts: alertsBatches[i],
		}

		buffers.AppendToSetBuffer(setBuffer, job)
	}

	logging.GetThenSendInfo(
		system.Logger,
		"dispatched jobs",
		func(event *logging.Event, level logging.Level) error {
			amounts := make([]int, len(alertsBatches))
			for i, alerts := range alertsBatches {
				amounts[i] = len(alerts)
			}

			logfmt.Unsigned(event, "jobs.returned_amount", setBuffer.Length)
			logfmt.Unsigneds(event, "jobs.ids", ids[:minLength]...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts[:minLength]...)

			return nil
		},
	)
}

func PutBusyJobs(system *DispatchSystem, jobs ...models.Job) {
	logging.GetThenSendDebug(
		system.Logger,
		"going to return jobs",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Integer(event, "jobs.requested_amount", len(jobs))

			return nil
		},
	)

	ids := make([]uint64, len(jobs))
	ids = models.JobsToIds(jobs, ids)
	pools.RemoveFromPool(system.AlertsPool, ids...)
	buffer.ResetAlertsInBuffer(system.AlertsBuffer, ids...)

	logging.GetThenSendInfo(
		system.Logger,
		"returned jobs",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", ids...)

			return nil
		},
	)
}
