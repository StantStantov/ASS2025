package main

import (
	"StantStantov/ASS/internal/agents"
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/dispatchers"
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func main() {
	ctx, stopCtx := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stopCtx()

	agentsAmount := 4
	chanceToCrash := 0.5

	logger := logging.NewLogger(os.Stdout, logfmt.MainFormat, logging.LevelDebug, 256)

	arrayPool := pools.NewArrayPool[agents.AgentId](uint64(agentsAmount))
	jobsPool := pools.NewArrayPool[models.Job](uint64(agentsAmount))
	jobPool := pools.NewJobPool(1)

	bufferSystem := buffer.NewBufferSystem(logger)
	dispatchSystem := dispatchers.NewDispatchSystem(
		bufferSystem,
		jobPool,
		logger,
	)
	agentSystem := agents.NewAgentSystem(
		uint64(agentsAmount),
		float32(chanceToCrash),
		bufferSystem,
		arrayPool,
		jobsPool,
		jobPool,
		logger,
	)

	msPerUpdate := 1.0
	previous := timeToFloat64(time.Now())
	lag := 0.0
	for {
		select {
		case <-ctx.Done():
			logging.GetThenSendDebug(
				logger,
				"stopped simulation",
				logging.NilFormat,
			)

			return
		default:
			current := timeToFloat64(time.Now())
			elapsed := current - previous
			previous = current
			lag += elapsed

			for lag >= msPerUpdate {
				agents.ProcessAgentSystem(agentSystem)
				dispatchers.ProcessDispatchSystem(dispatchSystem)

				lag -= msPerUpdate
			}
		}
	}
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
