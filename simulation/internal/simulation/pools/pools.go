package pools

import (
	"StantStantov/ASS/internal/simulation/models"
	"sync"

	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type PoolSystem struct {
	Queue   *doublyList
	Present *sparsemap.SparseMap[uint64, *poolNode]
	Locked  *sparseset.SparseSet[uint64]

	Logger *logging.Logger

	Mutex *sync.Mutex
}

func NewPoolSystem(
	capacity uint64,
	logger *logging.Logger,
) *PoolSystem {
	system := &PoolSystem{}

	system.Queue = &doublyList{}
	system.Present = sparsemap.NewSparseMap[uint64, *poolNode](capacity)
	system.Locked = sparseset.NewSparseSet(capacity)

	system.Mutex = &sync.Mutex{}

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

	minLength := bools.CountFalse[uint64](arePresent...)
	idsFiltered := make([]uint64, minLength)
	nodesToQueue := make([]*poolNode, minLength)
	currentAppendIndex := 0
	iterator := bools.IterOnlyFalse[uint64](arePresent...)
	for i := range iterator {
		job := jobs[i]
		jobId := job.Id
		node := &poolNode{
			Next:  nil,
			Prev:  nil,
			Value: jobId,
		}

		idsFiltered[currentAppendIndex] = jobId
		nodesToQueue[currentAppendIndex] = node
		currentAppendIndex++
	}

	pushNodesIntoDoublyList(system.Queue, nodesToQueue...)

	oksMove := make([]bool, len(nodesToQueue))
	oksMove = sparsemap.AddIntoSparseMap(system.Present, oksMove, idsFiltered, nodesToQueue)

	logging.GetThenSendInfo(
		system.Logger,
		"added new jobs into pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", idsFiltered...)

			return nil
		},
	)
}

func JobsPendingTotal(system *PoolSystem) uint64 {
	return system.Queue.Length
}

func JobsUnlockedTotal(system *PoolSystem) uint64 {
	return JobsPendingTotal(system) - JobsLockedTotal(system)
}

func JobsLockedTotal(system *PoolSystem) uint64 {
	return uint64(len(system.Locked.Dense))
}

func GetFromPool(system *PoolSystem, setBuffer []uint64) []uint64 {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	queue := system.Queue
	allIds := make([]uint64, queue.Length)
	currentNode := queue.Head
	for i := range allIds {
		id := currentNode.Value
		allIds[i] = id

		currentNode = currentNode.Next
	}

	arePresent := make([]bool, queue.Length)
	arePresent = sparseset.PresentInSparseSet(system.Locked, arePresent, allIds...)

	minLength := min(uint64(len(setBuffer)), bools.CountFalse[uint64](arePresent...))
	currentAppendIndex := uint64(0)
	iterator := bools.IterOnlyFalse[uint64](arePresent...)
	for i := range iterator {
		if currentAppendIndex == minLength {
			break
		}

		setBuffer[currentAppendIndex] = allIds[i]

		currentAppendIndex++
	}
	setBuffer = setBuffer[:minLength]

	oksAdded := make([]bool, minLength)
	oksAdded = sparseset.AddIntoSparseSet(system.Locked, oksAdded, setBuffer...)

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

	arePresent := make([]bool, len(ids))
	arePresent = sparseset.PresentInSparseSet(system.Locked, arePresent, ids...)

	minLength := bools.CountTrue[uint64](arePresent...)
	idsToRemove := make([]uint64, minLength)
	iterator := bools.IterOnlyTrue[uint64](arePresent...)
	for i := range iterator {
		idsToRemove[i] = ids[i]
	}

	nodesToRemove := make([]*poolNode, minLength)
	oksRemoved := make([]bool, minLength)
	nodesToRemove, oksRemoved = sparsemap.GetFromSparseMap(system.Present, nodesToRemove, oksRemoved, idsToRemove...)

	oksRemoved = sparsemap.RemoveFromSparseMap(system.Present, oksRemoved, idsToRemove...)

	oksRemoved = sparseset.RemoveFromSparseSet(system.Locked, oksRemoved, idsToRemove...)

	removeNodesFromDoublyList(system.Queue, nodesToRemove...)

	logging.GetThenSendInfo(
		system.Logger,
		"removed finished jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", ids...)

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
