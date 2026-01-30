package dispatchers

import (
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/pools"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type DispatchSystem struct {
	Buffer *buffer.BufferSystem

	JobPool *pools.JobPool

	Logger *logging.Logger
}

func NewDispatchSystem(
	buffer *buffer.BufferSystem,
	jobPool *pools.JobPool,
	logger *logging.Logger,
) *DispatchSystem {
	system := &DispatchSystem{}

	system.Buffer = buffer

	system.JobPool = jobPool

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "dispatch_system")
	})

	return system
}

func ProcessDispatchSystem(system *DispatchSystem) {
	jobs := buffer.DequeueAllFromBuffer(system.Buffer)

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

	pools.PutJobs(system.JobPool, jobs...)
}
