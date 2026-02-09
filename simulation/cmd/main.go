package main

import (
	"StantStantov/ASS/internal/agents"
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/commands"
	"StantStantov/ASS/internal/common/mempools"
	"StantStantov/ASS/internal/dispatchers"
	"StantStantov/ASS/internal/input"
	"StantStantov/ASS/internal/metrics"
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"
	"StantStantov/ASS/internal/render"
	"StantStantov/ASS/internal/responders"
	"StantStantov/ASS/internal/state"
	"os"
	"time"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func main() {
	state.InitState()

	commandsSystem := commands.NewCommandsSystem()
	inputSystem := input.NewInputSystem(commandsSystem)

	ui := &render.Ui{}
	render.InitUi(ui, inputSystem, commandsSystem)

	file, err := os.Create(".logs")
	if err != nil {
		panic(err)
	}

	agentsAmount := uint64(4)
	respondersAmount := uint64(2)
	chanceToCrash := float32(0.5)
	chanceToHandle := float32(0.5)

	agentsIdsPool := mempools.NewArrayPool[agents.AgentId](agentsAmount)
	respondersIdsPool := mempools.NewArrayPool[models.ResponderId](respondersAmount)
	jobsPool := mempools.NewArrayPool[models.Job](agentsAmount)

	logger := logging.NewLogger(
		file,
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

	go func() {
		msPerUpdate := 1.000
		previous := timeToFloat64(time.Now())
		lag := 0.0
		for {
			current := timeToFloat64(time.Now())
			elapsed := current - previous
			previous = current
			lag += elapsed

			commands.ProcessCommandsSystem(commandsSystem)
			for lag >= msPerUpdate {
				if !state.IsPaused {
					agents.ProcessAgentSystem(agentSystem)
					responders.ProcessRespondersSystem(respondersSystem)
					metrics.ProcessMetricsSystem(metricsSystem)
				}

				lag -= msPerUpdate
			}
		}
	}()

	render.RunUi(ui)
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
