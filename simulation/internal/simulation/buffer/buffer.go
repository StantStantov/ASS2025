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
	Values         *sparsemap.SparseMap[uint64, []models.MachineInfo]
	AlertsCapacity uint64

	Mutex *sync.Mutex

	Metrics *metrics.MetricsSystem
	Logger  *logging.Logger
}

func NewBufferSystem(
	capacity uint64,
	AlertsCapacity uint64,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *BufferSystem {
	system := &BufferSystem{}

	system.Values = sparsemap.NewSparseMap[uint64, []models.MachineInfo](capacity)
	system.AlertsCapacity = AlertsCapacity

	system.Mutex = &sync.Mutex{}

	system.Metrics = metrics
	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "buffer_system")
	})

	return system
}

func AddIntoBuffer(system *BufferSystem, ids []models.AgentId, alertsBatches [][]models.MachineInfo) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	minLength := min(len(ids), len(alertsBatches))
	values := make([][]models.MachineInfo, minLength)
	arePresent := make([]bool, minLength)
	values, arePresent = sparsemap.GetFromSparseMap(system.Values, values, arePresent, ids...)

	iterNewValues := bools.IterOnlyFalse[uint64](arePresent...)
	for i := range iterNewValues {
		values[i] = alertsBatches[i]
	}

	iterOldValues := bools.IterOnlyTrue[uint64](arePresent...)
	for i := range iterOldValues {
		newValue := alertsBatches[i]
		oldValue := values[i]

		oldValue = append(oldValue, newValue...)
		values[i] = oldValue
	}

	movedIntoBuffer := make([]bool, minLength)
	movedIntoBuffer = sparsemap.SaveIntoSparseMap(system.Values, movedIntoBuffer, ids, values)
	if bools.AnyFalse(movedIntoBuffer...) {
		panic(fmt.Sprintf("Save into Buffer %v %v", ids, movedIntoBuffer))
	}

	idsAdded := bools.CountFalse[uint64](arePresent...)
	alertsAdded := uint64(0)
	for _, alerts := range alertsBatches {
		alertsAdded += uint64(len(alerts))
	}

	metrics.AddToMetric(system.Metrics, metrics.JobsBufferedCounter, idsAdded)
	metrics.AddToMetric(system.Metrics, metrics.AlertsBufferedCounter, alertsAdded)

	logging.GetThenSendInfo(
		system.Logger,
		"added new alerts into buffer",
		func(event *logging.Event, level logging.Level) error {
			amounts := make([]int, len(alertsBatches))
			for i, alerts := range alertsBatches {
				amounts[i] = len(alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts...)

			return nil
		},
	)
}

func GetMultipleFromBuffer(system *BufferSystem, setBuffer *buffers.SetBuffer[[]models.MachineInfo, uint64], ids ...uint64) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	oksGet := make([]bool, len(setBuffer.Array))
	setBuffer.Array, oksGet = sparsemap.GetFromSparseMap(system.Values, setBuffer.Array, oksGet, ids...)
	setBuffer.Length += uint64(len(setBuffer.Array))

	logging.GetThenSendInfo(
		system.Logger,
		"got alerts from buffer",
		func(event *logging.Event, level logging.Level) error {
			alertsBatches := buffers.ValuesOfSetBuffer(setBuffer)
			amounts := make([]int, len(alertsBatches))
			for i, alerts := range alertsBatches {
				amounts[i] = len(alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts...)

			return nil
		},
	)
}
