package responders

import (
	"StantStantov/ASS/internal/dispatchers"
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/mempools"
	"iter"
	"math/rand"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type RespondersSystem struct {
	Responders        []models.ResponderId
	RespondersInfo    []models.ResponderInfo
	RespondersJob     []*models.Job
	MinChanceToHandle float32

	FreeResponders *responderList
	BusyResponders *responderList

	Dispatcher *dispatchers.DispatchSystem

	ArrayPool *mempools.ArrayPool[models.ResponderId]

	Logger *logging.Logger
}

func NewRespondersSystem(
	capacity uint64,
	minChanceToHandle float32,
	dispatcher *dispatchers.DispatchSystem,
	arrayPool *mempools.ArrayPool[models.ResponderId],
	logger *logging.Logger,
) *RespondersSystem {
	system := &RespondersSystem{}

	system.Responders = make([]models.ResponderId, capacity)
	system.RespondersInfo = make([]models.ResponderInfo, capacity)
	system.RespondersJob = make([]*models.Job, capacity)
	for i := range system.Responders {
		system.Responders[i] = models.ResponderId(i)
		system.RespondersInfo[i] = models.ResponderInfo{}
	}
	system.MinChanceToHandle = minChanceToHandle

	system.FreeResponders = newResponderList()
	system.BusyResponders = newResponderList()
	pushResponders(system.FreeResponders, system.Responders...)

	system.Dispatcher = dispatcher

	system.ArrayPool = arrayPool

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "responders_system")
	})

	return system
}

func ProcessRespondersSystem(system *RespondersSystem) {
	jobs := dispatchers.GetFreeJobs(system.Dispatcher, system.FreeResponders.Length)
	for _, job := range jobs {
		freeResponder, ok := popResponder(system.FreeResponders)
		if !ok {
			return
		}

		system.RespondersJob[freeResponder] = job

		pushResponders(system.BusyResponders, freeResponder)
	}

	freedResponders := mempools.GetArray(system.ArrayPool)
	stillBusyResponders := mempools.GetArray(system.ArrayPool)
	for id := range popAllresponders(system.BusyResponders) {
		responderJob := system.RespondersJob[id]

		currentChance := rand.Float32()
		if currentChance >= system.MinChanceToHandle {
			freedResponders = append(freedResponders, id)

			dispatchers.PutBusyJobs(system.Dispatcher, responderJob)
		} else {
			stillBusyResponders = append(stillBusyResponders, id)
		}
	}

	pushResponders(system.FreeResponders, freedResponders...)
	pushResponders(system.BusyResponders, stillBusyResponders...)

	logging.GetThenSendInfo(
		system.Logger,
		"polled responders for statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "responders.all.ids", system.Responders...)
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

func responders(list *responderList) iter.Seq[models.ResponderId] {
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
