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
	Values         *sparsemap.SparseMap[uint64, buffers.SetBuffer[models.MachineInfo, uint64]]
	AlertsCapacity uint64

	Mutex *sync.Mutex

	Metrics *metrics.MetricsSystem
	Logger  *logging.Logger
}

func NewBufferSystem(
	capacity uint64,
	alertsCapacity uint64,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *BufferSystem {
	system := &BufferSystem{}

	system.Values = sparsemap.NewSparseMap[uint64, buffers.SetBuffer[models.MachineInfo, uint64]](capacity)
	system.AlertsCapacity = alertsCapacity

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

	alertsTotal := uint64(0)
	alertsAdded := uint64(0)
	alertsSkipped := uint64(0)

	minLength := min(len(ids), len(alertsBatches))
	alertBuffers := make([]buffers.SetBuffer[models.MachineInfo, uint64], minLength)
	arePresent := make([]bool, minLength)
	alertBuffers, arePresent = sparsemap.GetFromSparseMap(system.Values, alertBuffers, arePresent, ids...)

	iterNewValues := bools.IterOnlyFalse[uint64](arePresent...)
	for i := range iterNewValues {
		alerts := alertsBatches[i]

		bufferNew := &alertBuffers[i]
		bufferNew.Array = make([]models.MachineInfo, system.AlertsCapacity)
		for _, alert := range alerts {
			buffers.AppendToSetBuffer(bufferNew, alert)

			alertsAdded++
		}
		alertsTotal += uint64(len(alerts))
	}

	iterOldValues := bools.IterOnlyTrue[uint64](arePresent...)
	for i := range iterOldValues {
		alerts := alertsBatches[i]

		bufferOld := &alertBuffers[i]
		for _, alert := range alerts {
			if bufferOld.Length != uint64(len(bufferOld.Array)) {
				buffers.AppendToSetBuffer(bufferOld, alert)
				alertsAdded++
			} else {
				alertsSkipped++
			}
		}
		alertsTotal += uint64(len(alerts))
	}

	movedIntoBuffer := make([]bool, minLength)
	movedIntoBuffer = sparsemap.SaveIntoSparseMap(system.Values, movedIntoBuffer, ids, alertBuffers)
	if bools.AnyFalse(movedIntoBuffer...) {
		panic(fmt.Sprintf("Save into Buffer %v %v", ids, movedIntoBuffer))
	}

	metrics.AddToMetric(system.Metrics, metrics.AlertsCounter, alertsAdded)
	metrics.AddToMetric(system.Metrics, metrics.AlertsBufferedCounter, alertsAdded)
	metrics.AddToMetric(system.Metrics, metrics.AlertsRewrittenCounter, alertsSkipped)

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

	oksGet := make([]bool, cap(setBuffer.Array))
	alertBuffers := make([]buffers.SetBuffer[models.MachineInfo, uint64], cap(setBuffer.Array))
	alertBuffers, oksGet = sparsemap.GetFromSparseMap(system.Values, alertBuffers, oksGet, ids...)

	for i := range alertBuffers {
		alertBuffer := &alertBuffers[i]
		alerts := buffers.ValuesOfSetBuffer(alertBuffer)
		buffers.AppendToSetBuffer(setBuffer, alerts)
	}

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
