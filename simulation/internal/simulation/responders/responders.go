package responders

import (
	"StantStantov/ASS/internal/common/mempools"
	"StantStantov/ASS/internal/simulation/dispatchers"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"fmt"
	"iter"
	"math/rand"

	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/collections/sparseset"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type RespondersSystem struct {
	Responders        []models.ResponderId
	RespondersInfo    []models.ResponderInfo
	RespondersJob     []models.Job
	MinChanceToHandle float32

	FreeResponders *sparseset.SparseSet[models.ResponderId]
	BusyResponders *sparseset.SparseSet[models.ResponderId]

	Dispatcher *dispatchers.DispatchSystem

	ArrayPool *mempools.ArrayPool[models.ResponderId]

	Metrics *metrics.MetricsSystem

	Logger *logging.Logger
}

func NewRespondersSystem(
	capacity uint64,
	minChanceToHandle float32,
	dispatcher *dispatchers.DispatchSystem,
	arrayPool *mempools.ArrayPool[models.ResponderId],
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *RespondersSystem {
	system := &RespondersSystem{}

	system.Responders = make([]models.ResponderId, capacity)
	system.RespondersInfo = make([]models.ResponderInfo, capacity)
	system.RespondersJob = make([]models.Job, capacity)
	for i := range system.Responders {
		system.Responders[i] = models.ResponderId(i)
		system.RespondersInfo[i] = models.ResponderInfo{}
	}
	system.MinChanceToHandle = minChanceToHandle

	system.FreeResponders = sparseset.NewSparseSet(capacity)
	system.BusyResponders = sparseset.NewSparseSet(capacity)

	oksAdded := make([]bool, len(system.Responders))
	oksAdded = sparseset.AddIntoSparseSet(system.FreeResponders, oksAdded, system.Responders...)

	system.Dispatcher = dispatcher

	system.ArrayPool = arrayPool

	system.Metrics = metrics

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "responders_system")
	})

	return system
}

func ProcessRespondersSystem(system *RespondersSystem) {
	freeResponders := make([]models.ResponderId, len(system.FreeResponders.Dense))
	copy(freeResponders, system.FreeResponders.Dense)

	jobs := make([]models.Job, len(freeResponders))
	jobs = dispatchers.GetFreeJobs(system.Dispatcher, jobs)

	minLength := min(len(freeResponders), len(jobs))
	freeRespondersToReceiveJobs := freeResponders[:minLength]
	jobs = jobs[:minLength]

	for i := range minLength {
		responder := freeRespondersToReceiveJobs[i]
		job := jobs[i]

		system.RespondersJob[responder] = job
	}

	oksRemoved := make([]bool, minLength)
	oksRemoved = sparseset.RemoveFromSparseSet(system.FreeResponders, oksRemoved, freeRespondersToReceiveJobs...)
	if !bools.AllTrue(oksRemoved...) {
		panic(fmt.Sprintf("Remove Busy %v %v", freeRespondersToReceiveJobs, oksRemoved))
	}

	oksAdded := make([]bool, minLength)
	oksAdded = sparseset.AddIntoSparseSet(system.BusyResponders, oksAdded, freeRespondersToReceiveJobs...)
	if !bools.AllTrue(oksAdded...) {
		panic(fmt.Sprintf("Add Busy %v %v", freeRespondersToReceiveJobs, oksAdded))
	}

	busyResponders := make([]models.ResponderId, len(system.BusyResponders.Dense))
	copy(busyResponders, system.BusyResponders.Dense)

	freedResponders := mempools.GetArray(system.ArrayPool)
	stillBusyResponders := mempools.GetArray(system.ArrayPool)
	for _, id := range busyResponders {
		responderJob := system.RespondersJob[id]

		currentChance := rand.Float32()
		if currentChance >= system.MinChanceToHandle {
			freedResponders = append(freedResponders, id)

			dispatchers.PutBusyJobs(system.Dispatcher, responderJob)
		} else {
			stillBusyResponders = append(stillBusyResponders, id)
		}
	}

	oksRemovedFreed := make([]bool, len(freedResponders))
	oksRemovedFreed = sparseset.RemoveFromSparseSet(system.BusyResponders, oksRemovedFreed, freedResponders...)
	if !bools.AllTrue(oksRemovedFreed...) {
		panic(fmt.Sprintf("Remove Freed %v %v", freedResponders, oksRemovedFreed))
	}

	oksAddedFreed := make([]bool, len(freedResponders))
	oksAddedFreed = sparseset.AddIntoSparseSet(system.FreeResponders, oksAddedFreed, freedResponders...)
	if !bools.AllTrue(oksAddedFreed...) {
		panic(fmt.Sprintf("Add Freed %v %v", freedResponders, oksAddedFreed))
	}

	metrics.AddToMetric(system.Metrics, metrics.RespondersFreeCounter, uint64(len(system.FreeResponders.Dense)))
	metrics.AddToMetric(system.Metrics, metrics.RespondersBusyCounter, uint64(len(system.BusyResponders.Dense)))

	logging.GetThenSendInfo(
		system.Logger,
		"polled responders for statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "responders.freed.ids", freedResponders...)
			logfmt.Unsigneds(event, "responders.still_busy.ids", stillBusyResponders...)

			return nil
		},
	)

	mempools.PutArrays(system.ArrayPool, freedResponders, stillBusyResponders)
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
