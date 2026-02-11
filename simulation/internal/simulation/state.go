package simulation

import (
	"StantStantov/ASS/internal/simulation/agents"
	"StantStantov/ASS/internal/simulation/buffer"
	"StantStantov/ASS/internal/simulation/commands"
	"StantStantov/ASS/internal/simulation/dispatchers"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/pools"
	"StantStantov/ASS/internal/simulation/responders"
	"time"

	"github.com/StantStantov/rps/swamp/logging"
)

var (
	Buffer           *buffer.BufferSystem         = nil
	Pool             *pools.PoolSystem            = nil
	CommandsSystem   *commands.CommandsSystem     = nil
	DispatchSystem   *dispatchers.DispatchSystem  = nil
	AgentsSystem     *agents.AgentSystem          = nil
	RespondersSystem *responders.RespondersSystem = nil
	MetricsSystem    *metrics.MetricsSystem       = nil

	IsPaused    bool   = true
	TickCounter uint64 = 0
)

func Init(
	agentsAmount uint64,
	respondersAmount uint64,
	chanceToCrash float32,
	chanceToHandle float32,
	logger *logging.Logger,
) {
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
	commandsSystem := commands.NewCommandsSystem()
	dispatchSystem := dispatchers.NewDispatchSystem(
		bufferSystem,
		poolSystem,
		metricsSystem,
		logger,
	)
	agentsSystem := agents.NewAgentSystem(
		agentsAmount,
		chanceToCrash,
		dispatchSystem,
		metricsSystem,
		logger,
	)
	respondersSystem := responders.NewRespondersSystem(
		respondersAmount,
		chanceToHandle,
		dispatchSystem,
		metricsSystem,
		logger,
	)

	Buffer = bufferSystem
	Pool = poolSystem
	CommandsSystem = commandsSystem
	DispatchSystem = dispatchSystem
	AgentsSystem = agentsSystem
	RespondersSystem = respondersSystem
	MetricsSystem = metricsSystem

	IsPaused = true
	TickCounter = 0
}

func RunEventLoop() {
	msPerUpdate := 1.000
	previous := TimeNowInSeconds()
	lag := 0.0
	for {
		current := TimeNowInSeconds()
		elapsed := current - previous
		previous = current
		lag += elapsed

		commands.ProcessCommandsSystem(CommandsSystem)
		for lag >= msPerUpdate {
			if !IsPaused {
				agents.ProcessAgentSystem(AgentsSystem)
				responders.ProcessRespondersSystem(RespondersSystem)
				metrics.ProcessMetricsSystem(MetricsSystem)
				TickCounter++
			}

			lag -= msPerUpdate
		}
	}
}

func TimeNowInSeconds() float64 {
	timestamp := time.Now()

	return float64(timestamp.UnixNano() / 1e9)
}
