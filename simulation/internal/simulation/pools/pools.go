package pools

import (
	ptime "StantStantov/ASS/internal/common/time"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"fmt"
	"sync"

	"github.com/StantStantov/rps/swamp/behaivors/buffers"
	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
	"github.com/StantStantov/rps/swamp/filters"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type PoolSystem struct {
	Queue      *doublyList
	Present    *sparsemap.SparseMap[uint64, *poolNode]
	Locked     *sparseset.SparseSet[uint64]
	Timestamps *sparsemap.SparseMap[uint64, float64]

	PoppedAmount    uint64
	SpentTimeInPool float64

	Mutex *sync.Mutex

	Metrics *metrics.MetricsSystem
	Logger  *logging.Logger
}

func NewPoolSystem(
	capacity uint64,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *PoolSystem {
	system := &PoolSystem{}

	system.Queue = &doublyList{}
	system.Present = sparsemap.NewSparseMap[uint64, *poolNode](capacity)
	system.Locked = sparseset.NewSparseSet(capacity)
	system.Timestamps = sparsemap.NewSparseMap[uint64, float64](capacity)

	system.Mutex = &sync.Mutex{}

	system.Metrics = metrics
	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "pool_system")
	})

	return system
}

func MoveIfNewIntoPool(system *PoolSystem, ids []models.AgentId) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	arePresent := make([]bool, len(ids))
	arePresent = sparsemap.PresentInSparseMap(system.Present, arePresent, ids...)

	idsNewAmount := bools.CountFalse[uint64](arePresent...)
	idsFiltered := make([]models.AgentId, idsNewAmount)
	idsBuffer := &buffers.SetBuffer[models.AgentId, uint64]{Array: idsFiltered}
	filters.KeepIfFalse(idsBuffer, ids, arePresent)

	nodesFiltered := make([]*poolNode, idsNewAmount)
	for i := range idsBuffer.Length {
		id := idsFiltered[i]
		nodesFiltered[i] = &poolNode{
			Next:  nil,
			Prev:  nil,
			Value: id,
		}
	}

	pushNodesIntoDoublyList(system.Queue, nodesFiltered...)

	movedIntoPool := make([]bool, len(nodesFiltered))
	movedIntoPool = sparsemap.AddIntoSparseMap(system.Present, movedIntoPool, idsFiltered, nodesFiltered)
	if bools.AnyFalse(movedIntoPool...) {
		panic(fmt.Sprintf("Add into Pool %v %v", idsFiltered, movedIntoPool))
	}

	metrics.AddToMetric(system.Metrics, metrics.JobsPendingCounter, idsNewAmount)
	metrics.AddToMetric(system.Metrics, metrics.JobsSkippedCounter, bools.CountTrue[uint64](arePresent...))

	logging.GetThenSendInfo(
		system.Logger,
		"added new jobs into pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", idsFiltered...)

			return nil
		},
	)
}

func GetFromPool(system *PoolSystem, setBuffer *buffers.SetBuffer[uint64, uint64]) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	logging.GetThenSendDebug(
		system.Logger,
		"going to get pending jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigned(event, "jobs.queued_amount", system.Queue.Length)
			logfmt.Unsigned(event, "jobs.locked_amount", sparseset.Length(system.Locked))
			logfmt.Integer(event, "jobs.requested_amount", len(setBuffer.Array))

			return nil
		},
	)

	queue := system.Queue
	allIds := make([]uint64, queue.Length)
	currentNode := queue.Head
	for i := range allIds {
		id := currentNode.Value
		allIds[i] = id

		currentNode = currentNode.Next
	}

	areLocked := make([]bool, queue.Length)
	areLocked = sparseset.PresentInSparseSet(system.Locked, areLocked, allIds...)

	filters.KeepIfFalse(setBuffer, allIds, areLocked)

	jobsToLockAmount := setBuffer.Length
	lockedJobs := make([]bool, jobsToLockAmount)
	lockedJobs = sparseset.AddIntoSparseSet(system.Locked, lockedJobs, setBuffer.Array...)
	if bools.AnyFalse(lockedJobs...) {
		panic(fmt.Sprintf("Lock Pool Jobs %v %v", setBuffer, lockedJobs))
	}

	addTime := ptime.TimeNowInSeconds()
	timestamps := make([]float64, jobsToLockAmount)
	for i := range timestamps {
		timestamps[i] = addTime
	}

	addTimestamps := make([]bool, jobsToLockAmount)
	sparsemap.SaveIntoSparseMap(system.Timestamps, addTimestamps, setBuffer.Array, timestamps)
	if bools.AnyFalse(lockedJobs...) {
		panic(fmt.Sprintf("Added Timestamps %v %v", setBuffer, lockedJobs))
	}

	metrics.AddToMetric(system.Metrics, metrics.JobsLockedCounter, jobsToLockAmount)

	logging.GetThenSendInfo(
		system.Logger,
		"got pending jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", setBuffer.Array...)

			return nil
		},
	)
}

func RemoveFromPool(system *PoolSystem, ids ...uint64) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	minLength := len(ids)
	nodes := make([]*poolNode, minLength)
	arePresent := make([]bool, minLength)
	nodes, arePresent = sparsemap.GetFromSparseMap(system.Present, nodes, arePresent, ids...)

	toRemoveAmount := bools.CountTrue[uint64](arePresent...)
	idsToRemove := make([]uint64, toRemoveAmount)
	idsBuffer := &buffers.SetBuffer[uint64, uint64]{Array: idsToRemove}
	filters.KeepIfTrue(idsBuffer, ids, arePresent)
	nodesToRemove := make([]*poolNode, toRemoveAmount)
	nodesBuffer := &buffers.SetBuffer[*poolNode, uint64]{Array: nodesToRemove}
	filters.KeepIfTrue(nodesBuffer, nodes, arePresent)

	removedFromPresent := make([]bool, toRemoveAmount)
	removedFromPresent = sparsemap.RemoveFromSparseMap(system.Present, removedFromPresent, idsToRemove...)
	if bools.AnyFalse(removedFromPresent...) {
		panic(fmt.Sprintf("Removed From Present %v %v", idsToRemove, removedFromPresent))
	}

	removedFromLocked := make([]bool, toRemoveAmount)
	removedFromLocked = sparseset.RemoveFromSparseSet(system.Locked, removedFromLocked, idsToRemove...)
	if bools.AnyFalse(removedFromLocked...) {
		panic(fmt.Sprintf("Removed From Locked %v %v", idsToRemove, removedFromLocked))
	}

	removeNodesFromDoublyList(system.Queue, nodesToRemove...)

	metrics.AddToMetric(system.Metrics, metrics.JobsUnlockedCounter, toRemoveAmount)

	getTimestamps := make([]bool, toRemoveAmount)
	timestampsPut := make([]float64, toRemoveAmount)
	sparsemap.GetFromSparseMap(system.Timestamps, timestampsPut, getTimestamps, idsToRemove...)
	if bools.AnyFalse(getTimestamps...) {
		panic(fmt.Sprintf("Get Timestamps %v %v", idsToRemove, getTimestamps))
	}

	timestampPopped := ptime.TimeNowInSeconds()
	timestampSpentInPool := float64(0)
	for i := range timestampsPut {
		timestampSpentInPool += timestampPopped - timestampsPut[i]
	}

	system.SpentTimeInPool += timestampSpentInPool
	system.PoppedAmount += toRemoveAmount

	logging.GetThenSendInfo(
		system.Logger,
		"removed finished jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", idsToRemove...)

			return nil
		},
	)
}

type doublyList struct {
	Head   *poolNode
	Tail   *poolNode
	Length uint64
}

type poolNode struct {
	Next  *poolNode
	Prev  *poolNode
	Value uint64
}

func pushNodesIntoDoublyList(list *doublyList, nodes ...*poolNode) {
	if list == nil || nodes == nil {
		return
	}

	for _, node := range nodes {
		if node == nil {
			continue
		}

		node.Next = nil
		if list.Tail == nil {
			list.Head = node
			list.Tail = node
			node.Prev = nil
		} else {
			node.Prev = list.Tail
			list.Tail.Next = node
			list.Tail = node
		}

		list.Length++
	}
}

func pushListIntoDoublyList(list *doublyList, listToPush *doublyList) {
	if list == nil || listToPush == nil || listToPush.Head == nil {
		return
	}

	if list.Tail == nil {
		list.Head = listToPush.Head
		list.Tail = listToPush.Tail
	} else {
		newNode := listToPush.Head
		newNode.Prev = list.Tail
		list.Tail.Next = newNode
		list.Tail = listToPush.Tail
	}

	list.Length += listToPush.Length
}

func removeNodesFromDoublyList(list *doublyList, nodes ...*poolNode) {
	if list == nil || nodes == nil {
		return
	}

	for _, node := range nodes {
		if node == nil {
			continue
		}

		prev := node.Prev
		next := node.Next

		if prev != nil {
			prev.Next = next
		} else {
			list.Head = next
		}

		if next != nil {
			next.Prev = prev
		} else {
			list.Tail = prev
		}

		list.Length--
	}
}
