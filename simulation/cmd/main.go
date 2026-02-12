package main

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/ui"
	"io"
	"os"
	"strings"
	"time"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func main() {
	file, err := os.Create(".logs")
	if err != nil {
		panic(err)
	}

	logsBuffer := &strings.Builder{}

	multiWriter := io.MultiWriter(file, logsBuffer)
	logger := logging.NewLogger(
		multiWriter,
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
	ui.Init(simulation.CommandsSystem, logsBuffer)

	go simulation.RunEventLoop()
	ui.RunEventLoop()
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
