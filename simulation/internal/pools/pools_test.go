package pools_test

import (
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"
	"fmt"
	"slices"
	"testing"
)

func TestPool(t *testing.T) {
	wantLength := uint64(4)
	pool := pools.NewPoolSystem(wantLength)

	wantJobs := make([]*models.Job, wantLength)
	for i := range wantJobs {
		wantJobs[i] = &models.Job{Id: uint64(i), Alerts: nil}
	}

	pools.MoveIfNewIntoPool(pool, wantJobs...)

	gotJobs := make([]*models.Job, wantLength)
	pools.GetFromPool(pool, gotJobs)

	gotLength := uint64(len(gotJobs))
	if wantLength != gotLength {
		t.Fatalf("%d != %d", wantLength, gotLength)
	}

	for _, gotJob := range gotJobs {
		contains := slices.ContainsFunc(wantJobs, func(e *models.Job) bool {
			return e.Id == gotJob.Id
		})
		if !contains {
			t.Fatalf("%v is not present", gotJob)
		}
	}

	fmt.Fprintln(t.Output(), pool.List.Length, pool.List.Head, pool.List.Tail)
	current := pool.List.Head
	for current != nil {
		fmt.Fprintf(t.Output(), "%p %p %p %v\n", current, current.Prev, current.Next, current.Value)

		current = current.Next
	}

	for _, entry := range pool.Locked.Dense {
		fmt.Fprintf(t.Output(), "%p\n", entry.Value)
	}

	pools.RemoveFromPool(pool, gotJobs...)

	fmt.Fprintln(t.Output(), pool.List.Length, pool.List.Head, pool.List.Tail)
	current = pool.List.Head
	for current != nil {
		fmt.Fprintf(t.Output(), "%p %p %p %v\n", current, current.Prev, current.Next, current.Value)

		current = current.Next
	}

	for _, entry := range pool.Unlocked.Dense {
		fmt.Fprintf(t.Output(), "%p\n", entry.Value)
	}
}
