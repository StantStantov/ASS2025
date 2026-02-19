package simulation

import (
	ptime "StantStantov/ASS/internal/common/time"
	"StantStantov/ASS/internal/simulation/agents"
	"StantStantov/ASS/internal/simulation/buffer"
	"StantStantov/ASS/internal/simulation/commands"
	"StantStantov/ASS/internal/simulation/dispatchers"
	"StantStantov/ASS/internal/simulation/framebuffer"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/pools"
	"StantStantov/ASS/internal/simulation/responders"

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

	Logbuffer *framebuffer.Buffer = nil

	MsPerUpdate float64 = 1.000
	IsPaused    bool    = true
	TickCounter uint64  = 0
)

func Init(
	msPerUpdate float64,
	agentsAmount uint64,
	respondersAmount uint64,
	chanceToCrash float32,
	alertsCapacity uint64,
	chanceToHandle float32,
	logbuffer *framebuffer.Buffer,
	logger *logging.Logger,
) {
	commandsSystem := commands.NewCommandsSystem()
	metricsSystem := metrics.NewMetricsSystem(
		logger,
	)
	bufferSystem := buffer.NewBufferSystem(
		agentsAmount,
		alertsCapacity,
		metricsSystem,
		logger,
	)
	poolSystem := pools.NewPoolSystem(
		agentsAmount,
		metricsSystem,
		logger,
	)
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

	Logbuffer = logbuffer

	MsPerUpdate = msPerUpdate
	IsPaused = true
	TickCounter = 0
}

func RunEventLoop() {
	previous := ptime.TimeNowInSeconds()
	lag := 0.0
	for {
		current := ptime.TimeNowInSeconds()
		elapsed := current - previous
		previous = current
		lag += elapsed

		commands.ProcessCommandsSystem(CommandsSystem)
		for lag >= MsPerUpdate {
			if !IsPaused {
				agents.ProcessAgentSystem(AgentsSystem)
				responders.ProcessRespondersSystem(RespondersSystem)
				framebuffer.Next(Logbuffer)
				TickCounter++
			}

			lag -= MsPerUpdate
		}
	}
}
