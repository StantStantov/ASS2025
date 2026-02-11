package responders

import (
	"StantStantov/ASS/internal/simulation/dispatchers"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"fmt"
	"iter"
	"math/rand"

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

	Logger *logging.Logger
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
	jobsToBusy = dispatchers.GetFreeJobs(system.Dispatcher, jobsToBusy)

	minLength := uint64(len(jobsToBusy))
	respondersToBusy := respondersFree[:minLength]

	removedFromFree := make([]bool, minLength)
	sparseset.RemoveFromSparseSet(system.Free, removedFromFree, respondersToBusy...)
	if !bools.AllTrue(removedFromFree...) {
		panic(fmt.Sprintf("Remove from Free %v %v", respondersToBusy, removedFromFree))
	}

	addedToBusy := make([]bool, minLength)
	addedToBusy = sparsemap.AddIntoSparseMap(system.Busy, addedToBusy, respondersToBusy, jobsToBusy)
	if !bools.AllTrue(addedToBusy...) {
		panic(fmt.Sprintf("Add to Busy %v %v", respondersToBusy, addedToBusy))
	}

	logging.GetThenSendInfo(
		system.Logger,
		"gave free responders new jobs",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "responders.ids", respondersToBusy...)

			return nil
		},
	)

	amountBusy := sparsemap.Length(system.Busy)
	areFreed := make([]bool, amountBusy)
	for i := range amountBusy {
		currentChance := rand.Float32()
		free := currentChance >= system.MinChanceToHandle
		areFreed[i] = free
	}

	amountFreed, amountStillBusy := bools.CountBools[models.ResponderId, models.ResponderId](areFreed...)
	respondersStillBusy := make([]models.ResponderId, amountStillBusy)
	respondersFreed := make([]models.ResponderId, amountFreed)
	respondersStillBusy, respondersFreed = filters.SeparateByBools(respondersStillBusy, respondersFreed, system.Responders, areFreed)

	jobsToFree := make([]models.Job, amountFreed)
	gotJobsToFree := make([]bool, amountFreed)
	jobsToFree, gotJobsToFree = sparsemap.GetFromSparseMap(system.Busy, jobsToFree, gotJobsToFree, respondersFreed...)
	if !bools.AllTrue(gotJobsToFree...) {
		panic(fmt.Sprintf("Get Jobs to Free %v %v", respondersFreed, gotJobsToFree))
	}

	dispatchers.PutBusyJobs(system.Dispatcher, jobsToFree...)

	oksRemovedFreed := make([]bool, amountFreed)
	oksRemovedFreed = sparsemap.RemoveFromSparseMap(system.Busy, oksRemovedFreed, respondersFreed...)
	if !bools.AllTrue(oksRemovedFreed...) {
		panic(fmt.Sprintf("Remove Freed %v %v", respondersFreed, oksRemovedFreed))
	}

	oksAddedFreed := make([]bool, amountFreed)
	oksAddedFreed = sparseset.AddIntoSparseSet(system.Free, oksAddedFreed, respondersFreed...)
	if !bools.AllTrue(oksAddedFreed...) {
		panic(fmt.Sprintf("Add Freed %v %v", respondersFreed, oksAddedFreed))
	}

	metrics.AddToMetric(system.Metrics, metrics.RespondersBusyCounter, sparsemap.Length(system.Busy))
	metrics.AddToMetric(system.Metrics, metrics.RespondersFreeCounter, sparseset.Length(system.Free))

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

type responderList struct {
	Head   *responderNode
	Tail   *responderNode
	Length uint64
}

type responderNode struct {
	Next  *responderNode
	Value models.ResponderId
}

func newResponderList() *responderList {
	list := &responderList{}

	node := &responderNode{Next: nil, Value: 0}
	list.Head = node
	list.Tail = node
	list.Length = 0

	return list
}

func pushResponders(list *responderList, responders ...models.ResponderId) {
	for _, id := range responders {
		node := &responderNode{Next: nil, Value: id}

		tail := list.Tail
		tail.Next = node
		list.Tail = node
	}
	list.Length += uint64(len(responders))
}

func popResponder(list *responderList) (models.ResponderId, bool) {
	if list.Length == 0 {
		return 0, false
	}

	freeId := uint64(0)

	head := list.Head
	tail := list.Tail
	next := head.Next
	for {
		if head == tail {
			if next == nil {
				return 0, false
			}

			tail.Next = next
		} else {
			freeId = next.Value

			list.Head = next
			list.Length--

			return freeId, true
		}
	}
}

func popAllresponders(list *responderList) iter.Seq[models.ResponderId] {
	return func(yield func(models.ResponderId) bool) {
		for {
			id, ok := popResponder(list)
			if !ok {
				return
			}

			if !yield(id) {
				return
			}
		}
	}
}

func IterResponders(list *responderList) iter.Seq[models.ResponderId] {
	return func(yield func(models.ResponderId) bool) {
		currentNode := list.Head
		for currentNode != nil {
			id := currentNode.Value
			if !yield(id) {
				return
			}

			currentNode = currentNode.Next
		}
	}
}
