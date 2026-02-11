package pools

import (
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"fmt"
	"sync"

	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
	"github.com/StantStantov/rps/swamp/filters"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type PoolSystem struct {
	Queue   *doublyList
	Present *sparsemap.SparseMap[uint64, *poolNode]
	Locked  *sparseset.SparseSet[uint64]

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

	system.Mutex = &sync.Mutex{}

	system.Metrics = metrics
	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "pool_system")
	})

	return system
}

func MoveIfNewIntoPool(system *PoolSystem, jobs ...models.Job) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	ids := make([]uint64, len(jobs))
	ids = models.JobsToIds(jobs, ids)

	arePresent := make([]bool, len(jobs))
	arePresent = sparsemap.PresentInSparseMap(system.Present, arePresent, ids...)

	jobsNewAmount := bools.CountFalse[uint64](arePresent...)
	jobsFiltered := make([]models.Job, jobsNewAmount)
	jobsFiltered = filters.KeepIfFalse(jobsFiltered, jobs, arePresent)

	idsFiltered := make([]uint64, jobsNewAmount)
	nodesFiltered := make([]*poolNode, jobsNewAmount)
	for i, job := range jobsFiltered {
		jobId := job.Id
		idsFiltered[i] = jobId
		nodesFiltered[i] = &poolNode{
			Next:  nil,
			Prev:  nil,
			Value: jobId,
		}
	}

	pushNodesIntoDoublyList(system.Queue, nodesFiltered...)

	movedIntoPool := make([]bool, len(nodesFiltered))
	movedIntoPool = sparsemap.AddIntoSparseMap(system.Present, movedIntoPool, idsFiltered, nodesFiltered)
	if !bools.AllTrue(movedIntoPool...) {
		panic(fmt.Sprintf("Add into Pool %v %v", idsFiltered, movedIntoPool))
	}

	metrics.AddToMetric(system.Metrics, metrics.JobsPendingCounter, jobsNewAmount)

	logging.GetThenSendInfo(
		system.Logger,
		"added new jobs into pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", idsFiltered...)

			return nil
		},
	)
}

func GetFromPool(system *PoolSystem, setBuffer []uint64) []uint64 {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	logging.GetThenSendDebug(
		system.Logger,
		"going to get pending jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigned(event, "jobs.queued_amount", system.Queue.Length)
			logfmt.Unsigned(event, "jobs.locked_amount", sparseset.Length(system.Locked))
			logfmt.Integer(event, "jobs.requested_amount", len(setBuffer))

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

	setBuffer = filters.KeepIfFalse(setBuffer, allIds, areLocked)

	jobsToLockAmount := uint64(len(setBuffer) )
	lockedJobs := make([]bool, jobsToLockAmount)
	lockedJobs = sparseset.AddIntoSparseSet(system.Locked, lockedJobs, setBuffer...)
	if !bools.AllTrue(lockedJobs...) {
		panic(fmt.Sprintf("Lock Pool Jobs %v %v", setBuffer, lockedJobs))
	}

	metrics.AddToMetric(system.Metrics, metrics.JobsLockedCounter, jobsToLockAmount)

	logging.GetThenSendInfo(
		system.Logger,
		"got pending jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", setBuffer...)

			return nil
		},
	)

	return setBuffer
}

func RemoveFromPool(system *PoolSystem, ids ...uint64) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	minLength := len(ids)
	nodesToRemove := make([]*poolNode, minLength)
	arePresent := make([]bool, minLength)
	nodesToRemove, arePresent = sparsemap.GetFromSparseMap(system.Present, nodesToRemove, arePresent, ids...)

	oksRemovedFromPresent := make([]bool, minLength)
	oksRemovedFromPresent = sparsemap.RemoveFromSparseMap(system.Present, oksRemovedFromPresent, ids...)

	oksRemovedFromLocked := make([]bool, minLength)
	oksRemovedFromLocked = sparseset.RemoveFromSparseSet(system.Locked, oksRemovedFromLocked, ids...)

	removeNodesFromDoublyList(system.Queue, nodesToRemove...)

	logging.GetThenSendInfo(
		system.Logger,
		"removed finished jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			idsRemoved := make([]uint64, minLength)
			idsRemoved = filters.KeepIfTrue(idsRemoved, ids, arePresent)
			logfmt.Unsigneds(event, "jobs.ids", idsRemoved...)

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
