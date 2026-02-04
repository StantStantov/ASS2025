package pools

import (
	"StantStantov/ASS/internal/models"
	"sync"

	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type PoolSystem struct {
	List     *doublyList
	Unlocked *sparsemap.SparseMap[uint64, *poolNode]
	Locked   *sparsemap.SparseMap[uint64, *poolNode]

	Logger *logging.Logger

	Mutex *sync.Mutex
}

func NewPoolSystem(
	capacity uint64,
	logger *logging.Logger,
) *PoolSystem {
	system := &PoolSystem{}

	system.List = &doublyList{}

	system.Unlocked = sparsemap.NewSparseMap[uint64, *poolNode](capacity)
	system.Locked = sparsemap.NewSparseMap[uint64, *poolNode](capacity)

	system.Mutex = &sync.Mutex{}

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "pool_system")
	})

	return system
}

func MoveIfNewIntoPool(system *PoolSystem, jobs ...*models.Job) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	listToPush := &doublyList{}

	jobsFiltered := make([]*models.Job, 0, len(jobs))
	nodesFiltered := make([]*poolNode, 0, len(jobs))
	for _, job := range jobs {
		existUnlocked := sparsemap.IsPresentInSparseMap(system.Unlocked, []bool{false}, job.Id)
		existLocked := sparsemap.IsPresentInSparseMap(system.Locked, []bool{false}, job.Id)
		if !existUnlocked[0] && !existLocked[0] {
			node := &poolNode{
				Next:  nil,
				Prev:  nil,
				Value: job,
			}
			pushNodeIntoDoublyList(listToPush, node)

			jobsFiltered = append(jobsFiltered, job)
			nodesFiltered = append(nodesFiltered, node)
		}
	}

	idsFiltered := make([]uint64, len(jobs))
	idsFiltered = models.JobsPtrToIds(jobs, idsFiltered)

	pushListIntoDoublyList(system.List, listToPush)

	oksMove := make([]bool, len(nodesFiltered))
	sparsemap.MoveIntoSparseMap(system.Unlocked, oksMove, idsFiltered, nodesFiltered)

	logging.GetThenSendInfo(
		system.Logger,
		"added new jobs into pool",
		func(event *logging.Event, level logging.Level) error {
			amounts := make([]int, len(jobsFiltered))
			for i, job := range jobsFiltered {
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", idsFiltered...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts...)

			return nil
		},
	)
}

func GetFromPool(system *PoolSystem, setBuffer []*models.Job) []*models.Job {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	ids := make([]uint64, len(setBuffer))
	nodes := make([]*poolNode, len(setBuffer))

	dense := system.Unlocked.Dense
	minLength := min(len(dense), len(setBuffer))
	for i := range minLength {
		entry := dense[i]
		entryNode := entry.Value
		job := entryNode.Value
		jobId := job.Id

		ids[i] = jobId
		nodes[i] = entryNode
		setBuffer[i] = job
	}

	oksRemove := make([]bool, minLength)
	sparsemap.RemoveFromSparseMap(system.Unlocked, oksRemove, ids...)
	oksMove := make([]bool, minLength)
	sparsemap.MoveIntoSparseMap(system.Locked, oksMove, ids, nodes)

	setBuffer = setBuffer[:minLength]

	logging.GetThenSendInfo(
		system.Logger,
		"got pending jobs from pool",
		func(event *logging.Event, level logging.Level) error {
			ids := make([]uint64, len(setBuffer))
			amounts := make([]int, len(setBuffer))
			for i, job := range setBuffer {
				ids[i] = job.Id
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts...)

			return nil
		},
	)

	return setBuffer
}

func RemoveFromPool(system *PoolSystem, jobs ...*models.Job) bool {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	ids := make([]uint64, len(jobs))
	ids = models.JobsPtrToIds(jobs, ids)

	nodes := make([]*poolNode, len(jobs))
	oksGet := make([]bool, len(ids))
	nodes, oksGet = sparsemap.GetFromSparseMap(system.Locked, nodes, oksGet, ids...)

	oksRemove := make([]bool, len(ids))
	sparsemap.RemoveFromSparseMap(system.Locked, oksRemove, ids...)

	removeFromDoublyList(system.List, nodes...)

	logging.GetThenSendInfo(
		system.Logger,
		"removed finished jobs from pool",
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

	return true
}

type doublyList struct {
	Head   *poolNode
	Tail   *poolNode
	Length uint64
}

type poolNode struct {
	Next  *poolNode
	Prev  *poolNode
	Value *models.Job
}

func pushNodeIntoDoublyList(list *doublyList, newNode *poolNode) {
	if list == nil || newNode == nil {
		return
	}

	if list.Tail == nil {
		list.Head = newNode
		list.Tail = newNode
		newNode.Prev = nil
	} else {
		newNode.Prev = list.Tail
		list.Tail.Next = newNode
		list.Tail = newNode
	}

	list.Length++
}

func pushListIntoDoublyList(list *doublyList, listToPush *doublyList) {
	if list == nil || listToPush == nil || listToPush.Head == nil || list.Tail == nil {
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

func removeFromDoublyList(list *doublyList, nodes ...*poolNode) {
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
	}

	list.Length -= uint64(len(nodes))
}
