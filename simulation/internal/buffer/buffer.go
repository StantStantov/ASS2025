package buffer

import (
	"StantStantov/ASS/internal/models"
	"sync"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type BufferSystem struct {
	Head   *alertBucket
	Length uint64

	Logger *logging.Logger

	Mutex sync.Mutex
}

type alertBucket struct {
	Next *alertBucket
	Job  models.Job
}

func NewBufferSystem(
	logger *logging.Logger,
) *BufferSystem {
	system := &BufferSystem{}

	bucket := &alertBucket{}
	system.Head = bucket
	system.Length = 0

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "buffer_system")
	})

	return system
}

func newAlertBucket(job models.Job) *alertBucket {
	bucket := &alertBucket{}

	bucket.Job = job

	return bucket
}

func EnqueueIntoBuffer(system *BufferSystem, jobs ...models.Job) {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	for _, job := range jobs {
		switch system.Length {
		case 0:
			system.Head = newAlertBucket(job)
		default:
			prevBucket := system.Head
			currentBucket := prevBucket
			for currentBucket != nil {
				if currentBucket.Job.Id == job.Id {
					currentBucket.Job.Alerts = append(currentBucket.Job.Alerts, job.Alerts...)

					return
				}

				prevBucket = currentBucket
				currentBucket = currentBucket.Next
			}

			prevBucket.Next = newAlertBucket(job)
		}
		system.Length++
	}

	logging.GetThenSendInfo(
		system.Logger,
		"enqueued into buffer new jobs",
		func(event *logging.Event, level logging.Level) error {
			ids:= make([]uint64, len(jobs))
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

func DequeueFromBuffer(system *BufferSystem) models.Job {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	if system.Length == 0 {
		return models.Job{}
	}

	job := system.Head.Job

	nextBucket := system.Head.Next
	system.Head = nextBucket
	system.Length--

	logging.GetThenSendInfo(
		system.Logger,
		"dequeued from buffer alerts",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "jobs.ids", job.Id)
			logfmt.Integers(event, "jobs.alerts.amounts", len(job.Alerts))
			return nil
		},
	)

	return job
}

func DequeueAllFromBuffer(system *BufferSystem) []models.Job {
	system.Mutex.Lock()
	defer system.Mutex.Unlock()

	if system.Length == 0 {
		return nil
	}

	allJobs := make([]models.Job, system.Length)
	for i := range system.Length {
		allJobs[i] = system.Head.Job

		nextBucket := system.Head.Next
		system.Head = nextBucket
	}
	system.Length = 0

	logging.GetThenSendInfo(
		system.Logger,
		"dequeued from buffer alerts",
		func(event *logging.Event, level logging.Level) error {
			ids:= make([]uint64, len(allJobs))
			amounts := make([]int, len(allJobs))
			for i, job := range allJobs {
				ids[i] = job.Id
				amounts[i] = len(job.Alerts)
			}

			logfmt.Unsigneds(event, "jobs.ids", ids...)
			logfmt.Integers(event, "jobs.alerts.amounts", amounts...)

			return nil
		},
	)

	return allJobs
}
