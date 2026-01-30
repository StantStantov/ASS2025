package buffer_test

import (
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/models"
	"slices"
	"sync"
	"testing"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func TestBuffer(t *testing.T) {
	logger := logging.NewLogger(
		t.Output(),
		logfmt.MainFormat,
		logging.LevelDebug,
		128,
	)

	t.Run("enqueue and dequeue", func(t *testing.T) {
		t.Parallel()

		wantLength := 32
		bufferSystem := buffer.NewBufferSystem(logger)
		multipleJobs := make([]models.Job, wantLength)
		for i := range wantLength {
			multipleJobs[i] = models.Job{Id: uint64(i), Alerts: make([]models.MachineInfo, 0)}
		}

		for _, job := range multipleJobs {
			buffer.EnqueueIntoBuffer(bufferSystem, job)
		}

		gotLength := int(bufferSystem.Length)
		if wantLength != gotLength {
			t.Fatalf("%d != %d", wantLength, gotLength)
		}

		current := bufferSystem.Head
		for current != nil {
			wantJob := multipleJobs[current.Job.Id]
			gotJob := buffer.DequeueFromBuffer(bufferSystem)
			if wantJob.Id != gotJob.Id {
				t.Fatalf("%v != %v", wantJob, gotJob)
			}

			current = current.Next
		}

		wantLength = 0
		gotLength = int(bufferSystem.Length)
		if wantLength != gotLength {
			t.Fatalf("%d != %d", wantLength, gotLength)
		}
	})
	t.Run("enqueue and dequeue all", func(t *testing.T) {
		t.Parallel()

		wantLength := 32
		bufferSystem := buffer.NewBufferSystem(logger)
		multipleJobs := make([]models.Job, wantLength)
		for i := range wantLength {
			multipleJobs[i] = models.Job{Id: uint64(i), Alerts: make([]models.MachineInfo, 0)}
		}

		for _, job := range multipleJobs {
			buffer.EnqueueIntoBuffer(bufferSystem, job)
		}

		gotLength := int(bufferSystem.Length)
		if wantLength != gotLength {
			t.Fatalf("%d != %d", wantLength, gotLength)
		}

		gotJobs := buffer.DequeueAllFromBuffer(bufferSystem)
		for i, gotJob := range gotJobs {
			wantJob := multipleJobs[i]
			if wantJob.Id != gotJob.Id {
				t.Fatalf("%v != %v", wantJob, gotJob)
			}
		}

		wantLength = 0
		gotLength = int(bufferSystem.Length)
		if wantLength != gotLength {
			t.Fatalf("%d != %d", wantLength, gotLength)
		}
	})
	t.Run("enqueue parallel", func(t *testing.T) {
		t.Parallel()

		wantLength := 32
		bufferSystem := buffer.NewBufferSystem(logger)
		multipleJobs := make([]models.Job, wantLength)
		for i := range wantLength {
			multipleJobs[i] = models.Job{Id: uint64(i), Alerts: make([]models.MachineInfo, 0)}
		}

		wg := &sync.WaitGroup{}
		for _, job := range multipleJobs {
			wg.Go(func() {
				buffer.EnqueueIntoBuffer(bufferSystem, job)
			})
		}
		wg.Wait()

		gotLength := int(bufferSystem.Length)
		if wantLength != gotLength {
			t.Fatalf("%d != %d", wantLength, gotLength)
		}

		gotJobs := buffer.DequeueAllFromBuffer(bufferSystem)
		for _, gotJob := range gotJobs {
			contains := slices.ContainsFunc(multipleJobs, func(e models.Job) bool {
				return e.Id == gotJob.Id
			})
			if !contains {
				t.Fatalf("%v is not present", gotJob)
			}
		}
	})
}
