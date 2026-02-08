package main

import (
	"StantStantov/ASS/internal/agents"
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/dispatchers"
	"StantStantov/ASS/internal/input"
	"StantStantov/ASS/internal/mempools"
	"StantStantov/ASS/internal/metrics"
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
	respondersAmount := uint64(2)
	chanceToCrash := float32(0.5)
	chanceToHandle := float32(0.5)

	agentsIdsPool := mempools.NewArrayPool[agents.AgentId](agentsAmount)
	respondersIdsPool := mempools.NewArrayPool[models.ResponderId](respondersAmount)
	jobsPool := mempools.NewArrayPool[models.Job](agentsAmount)

	inputSystem := input.NewInputSystem()

	logger := logging.NewLogger(
		os.Stdout,
		logfmt.MainFormat,
		logging.LevelDebug,
		256,
	)
	metricsSystem := metrics.NewMetricsSystem(
		logger,
	)
	bufferSystem := buffer.NewBufferSystem(
		agentsAmount,
		logger,
	)
	poolSystem := pools.NewPoolSystem(
		agentsAmount,
		logger,
	)
	dispatchSystem := dispatchers.NewDispatchSystem(
		bufferSystem,
		poolSystem,
		metricsSystem,
		logger,
	)
	agentSystem := agents.NewAgentSystem(
		agentsAmount,
		chanceToCrash,
		dispatchSystem,
		agentsIdsPool,
		jobsPool,
		metricsSystem,
		logger,
	)
	respondersSystem := responders.NewRespondersSystem(
		respondersAmount,
		chanceToHandle,
		dispatchSystem,
		respondersIdsPool,
		metricsSystem,
		logger,
	)

	input.ListenToInput(inputSystem)
	defer input.StopListening(inputSystem)

	msPerUpdate := 1.000
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

			input.ProcessInput(inputSystem)
			for lag >= msPerUpdate {
				agents.ProcessAgentSystem(agentSystem)
				responders.ProcessRespondersSystem(respondersSystem)
				metrics.ProcessMetricsSystem(metricsSystem)

				lag -= msPerUpdate
			}
		}
	}
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
