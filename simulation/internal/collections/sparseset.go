package collections

import (
	"sync"
)

type SparseSet[T ~uint64] struct {
	Dense  []T
	Sparse []uint64

	Mutex *sync.Mutex
}

func NewSparseSet[T ~uint64](capacity uint64) *SparseSet[T] {
	sparseMap := &SparseSet[T]{}

	sparseMap.Dense = make([]T, 0, capacity)
	sparseMap.Sparse = make([]uint64, capacity)
	sparseMap.Mutex = &sync.Mutex{}

	return sparseMap
}

func MoveIntoSparseSet[T ~uint64](sparseSet *SparseSet[T], handles ...T) bool {
	sparseSet.Mutex.Lock()
	defer sparseSet.Mutex.Unlock()

	for _, handle := range handles {
		index := sparseSet.Sparse[handle]
		currentLength := uint64(len(sparseSet.Dense))

		if index < currentLength && sparseSet.Dense[index] == handle {
			return false
		}

		sparseSet.Dense = append(sparseSet.Dense, handle)
		newLength := uint64(len(sparseSet.Dense))
		sparseSet.Sparse[handle] = newLength - 1
	}

	return true
}

func IsExistInSparseSet[T ~uint64](sparseSet *SparseSet[T], handle T) bool {
	index := sparseSet.Sparse[handle]
	currentLength := uint64(len(sparseSet.Dense))

	return index < currentLength && sparseSet.Dense[index] == handle
}

func RemoveFromSparseSet[T ~uint64](sparseSet *SparseSet[T], handles ...T) bool {
	sparseSet.Mutex.Lock()
	defer sparseSet.Mutex.Unlock()

	for _, handle := range handles {
		index := sparseSet.Sparse[handle]
		currentLength := uint64(len(sparseSet.Dense))

		if !(index < currentLength && sparseSet.Dense[index] == handle) {
			return false
		}

		handlePrevious := sparseSet.Dense[currentLength-1]
		sparseSet.Dense[index] = handlePrevious
		sparseSet.Sparse[handlePrevious] = index
		sparseSet.Dense = sparseSet.Dense[: currentLength-1 : cap(sparseSet.Dense)]
	}

	return true
}
