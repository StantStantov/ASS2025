package responders

import (
	"StantStantov/ASS/internal/simulation/dispatchers"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"fmt"
	"math/rand"

	"github.com/StantStantov/rps/swamp/behaivors/buffers"
	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
	"github.com/StantStantov/rps/swamp/filters"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type RespondersSystem struct {
	Responders        []models.ResponderId
	RespondersInfo    []models.ResponderInfo
	MinChanceToHandle float32

	Free *sparseset.SparseSet[models.ResponderId]
	Busy *sparsemap.SparseMap[models.ResponderId, models.Job]

	Dispatcher *dispatchers.DispatchSystem

	Metrics *metrics.MetricsSystem
	Logger  *logging.Logger
}

func NewRespondersSystem(
	capacity uint64,
	minChanceToHandle float32,
	dispatcher *dispatchers.DispatchSystem,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *RespondersSystem {
	system := &RespondersSystem{}

	system.Responders = make([]models.ResponderId, capacity)
	system.RespondersInfo = make([]models.ResponderInfo, capacity)
	for i := range system.Responders {
		system.Responders[i] = models.ResponderId(i)
		system.RespondersInfo[i] = models.ResponderInfo{}
	}
	system.MinChanceToHandle = minChanceToHandle

	system.Free = sparseset.NewSparseSet(capacity)
	system.Busy = sparsemap.NewSparseMap[models.ResponderId, models.Job](capacity)

	oksAdded := make([]bool, len(system.Responders))
	oksAdded = sparseset.AddIntoSparseSet(system.Free, oksAdded, system.Responders...)

	system.Dispatcher = dispatcher

	system.Metrics = metrics
	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "responders_system")
	})

	return system
}

func ProcessRespondersSystem(system *RespondersSystem) {
	amountFree := sparseset.Length(system.Free)
	respondersFree := make([]models.ResponderId, amountFree)
	respondersFree = sparseset.GetAllFromSparseSet(system.Free, respondersFree)

	jobsToBusy := make([]models.Job, amountFree)
	jobsToBusyBuffer := &buffers.SetBuffer[models.Job, uint64]{Array: jobsToBusy}
	dispatchers.GetFreeJobs(system.Dispatcher, jobsToBusyBuffer)

	minLength := uint64(len(jobsToBusy))
	respondersToBusy := respondersFree[:minLength]

	removedFromFree := make([]bool, minLength)
	removedFromFree = sparseset.RemoveFromSparseSet(system.Free, removedFromFree, respondersToBusy...)
	if bools.AnyFalse(removedFromFree...) {
		panic(fmt.Sprintf("Remove Busyed from Free %v %v", respondersToBusy, removedFromFree))
	}

	addedToBusy := make([]bool, minLength)
	addedToBusy = sparsemap.AddIntoSparseMap(system.Busy, addedToBusy, respondersToBusy, jobsToBusy)
	if bools.AnyFalse(addedToBusy...) {
		panic(fmt.Sprintf("Add Busyed to Busy %v %v", respondersToBusy, addedToBusy))
	}

	logging.GetThenSendInfo(
		system.Logger,
		"gave free responders new jobs",
		func(event *logging.Event, level logging.Level) error {
			jobsIds := make([]uint64, len(jobsToBusy))
			jobsIds = models.JobsToIds(jobsToBusy, jobsIds)

			logfmt.Unsigneds(event, "responders.ids", respondersToBusy...)
			logfmt.Unsigneds(event, "jobs.ids", jobsIds...)

			return nil
		},
	)

	amountBusy := sparsemap.Length(system.Busy)
	idsBusy := make([]models.ResponderId, amountBusy)
	idsBusy = sparsemap.GetAllKeysFromSparseMap(system.Busy, idsBusy)

	areFreed := make([]bool, amountBusy)
	for i := range amountBusy {
		currentChance := rand.Float32()
		free := currentChance >= system.MinChanceToHandle
		areFreed[i] = free
	}

	amountFreed, amountStillBusy := bools.CountBools[models.ResponderId, models.ResponderId](areFreed...)
	respondersStillBusy := make([]models.ResponderId, amountStillBusy)
	respondersFreed := make([]models.ResponderId, amountFreed)
	stillBusyBuffer := &buffers.SetBuffer[models.ResponderId, uint64]{Array: respondersStillBusy}
	freedBuffer := &buffers.SetBuffer[models.ResponderId, uint64]{Array: respondersFreed}
	filters.SeparateByBools(stillBusyBuffer, freedBuffer, idsBusy, areFreed)

	jobsToFree := make([]models.Job, len(respondersFreed))
	gotJobsToFree := make([]bool, len(respondersFreed))
	jobsToFree, gotJobsToFree = sparsemap.GetFromSparseMap(system.Busy, jobsToFree, gotJobsToFree, respondersFreed...)
	if bools.AnyFalse(gotJobsToFree...) {
		panic(fmt.Sprintf("Get Jobs to Free %v %v %v", system.Busy.Dense, respondersFreed, gotJobsToFree))
	}

	dispatchers.PutBusyJobs(system.Dispatcher, jobsToFree...)

	oksRemovedFreed := make([]bool, len(respondersFreed))
	oksRemovedFreed = sparsemap.RemoveFromSparseMap(system.Busy, oksRemovedFreed, respondersFreed...)
	if bools.AnyFalse(oksRemovedFreed...) {
		panic(fmt.Sprintf("Remove Freed From Busy %v %v", respondersFreed, oksRemovedFreed))
	}

	oksAddedFreed := make([]bool, len(respondersFreed))
	oksAddedFreed = sparseset.AddIntoSparseSet(system.Free, oksAddedFreed, respondersFreed...)
	if bools.AnyFalse(oksAddedFreed...) {
		panic(fmt.Sprintf("Add Freed To Free %v %v", respondersFreed, oksAddedFreed))
	}

	metrics.AddToMetric(system.Metrics, metrics.RespondersFreeCounter, FreeAmount(system))
	metrics.AddToMetric(system.Metrics, metrics.RespondersBusyCounter, BusyAmount(system))

	logging.GetThenSendInfo(
		system.Logger,
		"polled responders for statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "responders.freed.ids", respondersFreed...)
			logfmt.Unsigneds(event, "responders.still_busy.ids", respondersStillBusy...)

			return nil
		},
	)
}
