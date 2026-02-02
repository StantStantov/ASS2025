package buffer

import (
	"StantStantov/ASS/internal/collections"
	"StantStantov/ASS/internal/models"
	"sync"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type BufferSystem struct {
	Head   *alertBucket
	Length uint64

	Free *collections.SparseSet[uint64]
	Busy *collections.SparseSet[uint64]

	Logger *logging.Logger

	Mutex *sync.Mutex
}

type alertBucket struct {
	Next *alertBucket
	Job  *models.Job
}

func NewBufferSystem(
	capacity uint64,
	logger *logging.Logger,
) *BufferSystem {
	system := &BufferSystem{}

	bucket := &alertBucket{}
	system.Head = bucket
	system.Length = 0

	system.Free = collections.NewSparseSet[uint64](capacity)
	system.Busy = collections.NewSparseSet[uint64](capacity)

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "buffer_system")
	})

	system.Mutex = &sync.Mutex{}

	return system
}

func newAlertBucket(job *models.Job) *alertBucket {
	bucket := &alertBucket{}

	bucket.Job = job

	return bucket
}

func EnqueueIntoBuffer(system *BufferSystem, jobs ...*models.Job) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	for _, job := range jobs {
		switch system.Length {
		case 0:
			system.Head = newAlertBucket(job)
			collections.MoveIntoSparseSet(system.Free, job.Id)
			system.Length++
		default:
			isOld := false

			prevBucket := system.Head
			currentBucket := system.Head
			for currentBucket != nil {
				if currentBucket.Job.Id == job.Id {
					currentBucket.Job.Alerts = append(currentBucket.Job.Alerts, job.Alerts...)

					isOld = true

					break
				}

				prevBucket = currentBucket
				currentBucket = currentBucket.Next
			}

			if !isOld {
				prevBucket.Next = newAlertBucket(job)
				collections.MoveIntoSparseSet(system.Free, job.Id)
				system.Length++
			}
		}
	}

	logBufferEnque(system.Logger, jobs...)
}

func GetMultipleFreeFromBuffer(system *BufferSystem, maxAmount uint64) ([]*models.Job, bool) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	if system.Length == 0 {
		return nil, false
	}

	minLength := min(system.Length, uint64(len(system.Free.Dense)), maxAmount)
	ids := make([]uint64, 0, minLength)
	allJobs := make([]*models.Job, 0, minLength)
	currentBucket := system.Head
	for currentBucket != nil {
		currentJob := currentBucket.Job
		currentId := currentJob.Id

		exist := collections.IsExistInSparseSet(system.Free, currentId)
		if exist {
			ids = append(ids, currentId)
			allJobs = append(allJobs, currentJob)
		}
		if len(allJobs) == int(minLength) {
			break
		}

		currentBucket = currentBucket.Next
	}

	collections.RemoveFromSparseSet(system.Free, ids...)
	collections.MoveIntoSparseSet(system.Busy, ids...)

	logBufferGet(system.Logger, allJobs...)

	return allJobs, true
}

func PutMultipleBusyIntoBuffer(system *BufferSystem, jobs ...*models.Job) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	minLength := len(jobs)
	ids := make([]uint64, minLength)
	for i, job := range jobs {
		ids[i] = job.Id
	}

	collections.RemoveFromSparseSet(system.Busy, ids...)
	collections.MoveIntoSparseSet(system.Free, ids...)

	logBufferPut(system.Logger, jobs...)
}

func GetMultipleFromBuffer(system *BufferSystem, maxAmount uint64) ([]*models.Job, bool) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	if system.Length == 0 {
		return nil, false
	}

	minLength := system.Length
	ids := make([]uint64, 0, minLength)
	allJobs := make([]*models.Job, 0, system.Length)
	currentBucket := system.Head
	for currentBucket != nil {
		currentJob := currentBucket.Job
		currentId := currentJob.Id

		ids = append(ids, currentId)
		allJobs = append(allJobs, currentJob)
		if len(allJobs) == int(minLength) {
			break
		}

		currentBucket = currentBucket.Next
	}

	logBufferGet(system.Logger, allJobs...)

	return allJobs, true
}

func logBufferEnque(logger *logging.Logger, jobs ...*models.Job) {
	logging.GetThenSendInfo(
		logger,
		"enqueued new alerts into buffer",
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
}

func logBufferGet(logger *logging.Logger, jobs ...*models.Job) {
	logging.GetThenSendInfo(
		logger,
		"got alerts from buffer",
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
}

func logBufferPut(logger *logging.Logger, jobs ...*models.Job) {
	logging.GetThenSendInfo(
		logger,
		"put alerts into buffer",
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
}
