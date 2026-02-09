package buffer

import (
	"StantStantov/ASS/internal/simulation/models"
	"sync"

	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type BufferSystem struct {
	Values *sparsemap.SparseMap[uint64, models.Job]

	Logger *logging.Logger

	Mutex *sync.Mutex
}

func NewBufferSystem(
	capacity uint64,
	logger *logging.Logger,
) *BufferSystem {
	system := &BufferSystem{}

	system.Values = sparsemap.NewSparseMap[uint64, models.Job](capacity)

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "buffer_system")
	})

	system.Mutex = &sync.Mutex{}

	return system
}

func AddIntoBuffer(system *BufferSystem, jobs ...models.Job) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	ids := make([]uint64, len(jobs))
	ids = models.JobsToIds(jobs, ids)

	values := make([]models.Job, len(jobs))
	oksGet := make([]bool, len(jobs))
	values, oksGet = sparsemap.GetFromSparseMap(system.Values, values, oksGet, ids...)
	for i, ok := range oksGet {
		newValue := jobs[i]
		oldValue := values[i]
		if !ok {
			values[i] = newValue

			continue
		}

		if oldValue.Id == newValue.Id {
			oldValue.Alerts = append(oldValue.Alerts, newValue.Alerts...)

			values[i] = oldValue
		}
	}

	oksMove := make([]bool, len(jobs))
	oksMove = sparsemap.SaveIntoSparseMap(system.Values, oksMove, ids, values)

	logBufferAdd(system.Logger, jobs...)
}

func JobsTotal(system *BufferSystem) uint64 {
	return uint64(len(system.Values.Dense))
}

func AlertsTotal(system *BufferSystem) uint64 {
	total := uint64(0)
	values := system.Values.Dense
	for _, value := range values {
		job := value.Value
		alerts := job.Alerts

		total += uint64(len(alerts))
	}

	return total
}

func GetMultipleFromBuffer(system *BufferSystem, setBuffer []models.Job, ids ...uint64) []models.Job {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	oksGet := make([]bool, len(setBuffer))
	setBuffer, oksGet = sparsemap.GetFromSparseMap(system.Values, setBuffer, oksGet, ids...)

	logBufferGet(system.Logger, setBuffer...)

	return setBuffer
}

func logBufferAdd(logger *logging.Logger, jobs ...models.Job) {
	logging.GetThenSendInfo(
		logger,
		"added new alerts into buffer",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(jobs))
			amounts := make([]int, len(jobs))
			for i, job := range jobs {
				ids[i] = job.Id
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts...)

			return nil
		},
	)
}

func logBufferGet(logger *logging.Logger, jobs ...models.Job) {
	logging.GetThenSendInfo(
		logger,
		"got alerts from buffer",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(jobs))
			amounts := make([]int, len(jobs))
			for i, job := range jobs {
				ids[i] = job.Id
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts...)

			return nil
		},
	)
}
