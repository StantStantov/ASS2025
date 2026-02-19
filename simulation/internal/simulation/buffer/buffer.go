package buffer

import (
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"fmt"
	"sync"

	"github.com/StantStantov/rps/swamp/behaivors/buffers"
	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type BufferSystem struct {
	Values *sparsemap.SparseMap[uint64, models.Job]

	Mutex *sync.Mutex

	Metrics *metrics.MetricsSystem
	Logger  *logging.Logger
}

func NewBufferSystem(
	capacity uint64,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *BufferSystem {
	system := &BufferSystem{}

	system.Values = sparsemap.NewSparseMap[uint64, models.Job](capacity)

	system.Mutex = &sync.Mutex{}

	system.Metrics = metrics
	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "buffer_system")
	})

	return system
}

func AddIntoBuffer(system *BufferSystem, jobs ...models.Job) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	ids := make([]uint64, len(jobs))
	ids = models.JobsToIds(jobs, ids)

	values := make([]models.Job, len(jobs))
	arePresent := make([]bool, len(jobs))
	values, arePresent = sparsemap.GetFromSparseMap(system.Values, values, arePresent, ids...)

	iterNewValues := bools.IterOnlyFalse[uint64](arePresent...)
	for i := range iterNewValues {
		values[i] = jobs[i]
	}

	iterOldValues := bools.IterOnlyTrue[uint64](arePresent...)
	for i := range iterOldValues {
		newValue := jobs[i]
		oldValue := values[i]

		oldValue.Alerts = append(oldValue.Alerts, newValue.Alerts...)
		values[i] = oldValue
	}

	movedIntoBuffer := make([]bool, len(jobs))
	movedIntoBuffer = sparsemap.SaveIntoSparseMap(system.Values, movedIntoBuffer, ids, values)
	if bools.AnyFalse(movedIntoBuffer...) {
		panic(fmt.Sprintf("Save into Buffer %v %v", ids, movedIntoBuffer))
	}

	idsAdded := bools.CountTrue[uint64](arePresent...)
	alertsAdded := uint64(0)
	for _, job := range jobs {
		alertsAdded += uint64(len(job.Alerts))
	}

	metrics.AddToMetric(system.Metrics, metrics.JobsBufferedCounter, idsAdded)
	metrics.AddToMetric(system.Metrics, metrics.AlertsBufferedCounter, alertsAdded)

	logging.GetThenSendInfo(
		system.Logger,
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

func GetMultipleFromBuffer(system *BufferSystem, setBuffer *buffers.SetBuffer[models.Job, uint64], ids ...uint64) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	oksGet := make([]bool, len(setBuffer.Array))
	setBuffer.Array, oksGet = sparsemap.GetFromSparseMap(system.Values, setBuffer.Array, oksGet, ids...)

	logging.GetThenSendInfo(
		system.Logger,
		"got alerts from buffer",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, setBuffer.Length)
			ids = models.JobsToIds(setBuffer.Array, ids)

			logfmt.Unsigneds(event, "jobs.ids", ids...)

			return nil
		},
	)
}
