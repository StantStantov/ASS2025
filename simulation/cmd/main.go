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

	agentsAmount := uint64(4)
	respondersAmount := uint64(4)
	chanceToCrash := 0.5
	chanceToHandle := 0.5

	logger := logging.NewLogger(os.Stdout, logfmt.MainFormat, logging.LevelDebug, 256)

	agentsIdsPool := pools.NewArrayPool[agents.AgentId](agentsAmount)
	respondersIdsPool := pools.NewArrayPool[models.ResponderId](respondersAmount)
	jobsPool := pools.NewArrayPool[*models.Job](agentsAmount)
	jobPool := pools.NewJobPool(1)

	bufferSystem := buffer.NewBufferSystem(agentsAmount, logger)
	dispatchSystem := dispatchers.NewDispatchSystem(
		bufferSystem,
		logger,
	)

	agentSystem := agents.NewAgentSystem(
		agentsAmount,
		float32(chanceToCrash),
		dispatchSystem,
		agentsIdsPool,
		jobsPool,
		jobPool,
		logger,
	)
	respondersSystem := responders.NewRespondersSystem(
		respondersAmount,
		float32(chanceToHandle),
		dispatchSystem,
		respondersIdsPool,
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
				responders.ProcessRespondersSystem(respondersSystem)

				buffer.GetMultipleFromBuffer(bufferSystem, bufferSystem.Length)

				lag -= msPerUpdate
			}

		}
	}
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
