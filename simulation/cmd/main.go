package main

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/ui/input"
	"StantStantov/ASS/internal/ui/render"
	"os"
	"time"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func main() {
	file, err := os.Create(".logs")
	if err != nil {
		panic(err)
	}
	logger := logging.NewLogger(
		file,
		logfmt.MainFormat,
		logging.LevelDebug,
		256,
	)

	agentsAmount := uint64(4)
	respondersAmount := uint64(2)
	chanceToCrash := float32(0.5)
	chanceToHandle := float32(0.5)

	simulation.Init(
		agentsAmount,
		respondersAmount,
		chanceToCrash,
		chanceToHandle,
		logger,
	)

	inputSystem := input.NewInputSystem(simulation.CommandsSystem)
	ui := &render.Ui{}
	render.InitUi(ui, inputSystem, simulation.CommandsSystem)

	go simulation.RunEventLoop()

	render.RunUi(ui)
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
