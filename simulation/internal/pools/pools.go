package pools

import (
	"StantStantov/ASS/internal/models"
	"sync"
)

type ArrayPool[T any] struct {
	Arrays      *sync.Pool
	MaxCapacity uint64
}

func NewArrayPool[T any](maxCapacity uint64) *ArrayPool[T] {
	arrayPool := &ArrayPool[T]{}

	arrayPool.Arrays = &sync.Pool{
		New: func() any {
			return newArray[T](maxCapacity)
		},
	}
	arrayPool.MaxCapacity = maxCapacity

	return arrayPool
}

func GetArray[T any](pool *ArrayPool[T]) []T {
	got := pool.Arrays.Get()
	array, ok := got.([]T)
	if !ok {
		return nil
	}

	return array
}

func PutArrays[T any](pool *ArrayPool[T], arrays ...[]T) {
	for _, array := range arrays {
		array = array[:0:pool.MaxCapacity]
		pool.Arrays.Put(array)
	}
}

type JobPool struct {
	Alerts            *sync.Pool
	MinAlertsCapacity uint64
}

func NewJobPool(minAlertsCapacity uint64) *JobPool {
	jobPool := &JobPool{}

	jobPool.Alerts = &sync.Pool{
		New: func() any {
			return newArray[models.MachineInfo](minAlertsCapacity)
		},
	}
	jobPool.MinAlertsCapacity = minAlertsCapacity

	return jobPool
}

func GetJob(pool *JobPool) models.Job {
	got := pool.Alerts.Get()
	alerts, ok := got.([]models.MachineInfo)
	if !ok {
		return models.Job{}
	}

	job := models.Job{
		Id:     0,
		Alerts: alerts,
	}

	return job
}

func PutJobs(pool *JobPool, jobs ...models.Job) {
	for _, job := range jobs {
		job.Alerts = job.Alerts[:0:pool.MinAlertsCapacity]
		pool.Alerts.Put(job.Alerts)
	}
}

func newArray[T any](capacity uint64) []T {
	return make([]T, 0, capacity)
}
