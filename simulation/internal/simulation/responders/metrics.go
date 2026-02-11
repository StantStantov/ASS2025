package responders

import (
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
)

func FreeAmount(system *RespondersSystem) uint64 {
	return sparseset.Length(system.Free)
}

func BusyAmount(system *RespondersSystem) uint64 {
	return sparsemap.Length(system.Busy)
}
