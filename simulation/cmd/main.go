package main

import (
	"StantStantov/ASS/internal/agents"
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/dispatchers"
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"
	"StantStantov/ASS/internal/responders"
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
	respondersAmount := 4
	chanceToCrash := 0.5
	chanceToHandle := 0.5

	logger := logging.NewLogger(os.Stdout, logfmt.MainFormat, logging.LevelDebug, 256)

	agentsIdsPool := pools.NewArrayPool[agents.AgentId](uint64(agentsAmount))
	respondersIdsPool := pools.NewArrayPool[models.ResponderId](uint64(respondersAmount))
	jobsPool := pools.NewArrayPool[models.Job](uint64(agentsAmount))
	jobPool := pools.NewJobPool(1)

	bufferSystem := buffer.NewBufferSystem(logger)

	respondersSystem := responders.NewRespondersSystem(
		uint64(agentsAmount),
		float32(chanceToHandle),
		respondersIdsPool,
		jobPool,
		logger,
	)
	agentSystem := agents.NewAgentSystem(
		uint64(agentsAmount),
		float32(chanceToCrash),
		bufferSystem,
		agentsIdsPool,
		jobsPool,
		jobPool,
		logger,
	)
	dispatchSystem := dispatchers.NewDispatchSystem(
		bufferSystem,
		respondersSystem,
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
				responders.ProcessRespondersSystem(respondersSystem)

				lag -= msPerUpdate
			}
		}
	}
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
