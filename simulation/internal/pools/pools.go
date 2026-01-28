package pools

import (
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

func newArray[T any](capacity uint64) []T {
	return make([]T, 0, capacity)
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
