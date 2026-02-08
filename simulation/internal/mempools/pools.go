package mempools

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

type ValuePool[T any] struct {
	Values *sync.Pool
}

func NewValuePool[T any](newFunc func() T) *ValuePool[T] {
	valuePool := &ValuePool[T]{}

	valuePool.Values = &sync.Pool{
		New: func() any {
			return newFunc()
		},
	}

	return valuePool
}

func GetValueFromPool[T any](pool *ValuePool[T]) T {
	got := pool.Values.Get()
	value, ok := got.(T)
	if !ok {
		var zero T

		return zero
	}

	return value
}

func PutValueIntoPool[T any](pool *ValuePool[T], jobs ...T) {
	for _, job := range jobs {
		pool.Values.Put(job)
	}
}

func newArray[T any](capacity uint64) []T {
	return make([]T, 0, capacity)
}
