package main

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/framebuffer"
	"StantStantov/ASS/internal/ui"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func main() {
	logBuffer := &framebuffer.Buffer{}
	framebuffer.InitBuffer(logBuffer)

	logger := logging.NewLogger(
		logBuffer,
		logfmt.MainFormat,
		logging.LevelDebug,
		256,
	)

	msPerUpdate := float64(1.000)
	agentsAmount := uint64(8)
	respondersAmount := uint64(4)
	chanceToCrash := float32(0.5)
	chanceToHandle := float32(0.5)

	simulation.Init(
		msPerUpdate,
		agentsAmount,
		respondersAmount,
		chanceToCrash,
		chanceToHandle,
		logBuffer,
		logger,
	)
	ui.Init(simulation.CommandsSystem, logBuffer)

	go simulation.RunEventLoop()
	ui.RunEventLoop()
}
