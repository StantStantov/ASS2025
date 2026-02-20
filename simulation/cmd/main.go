package main

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/framebuffer"
	"StantStantov/ASS/internal/ui"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func main() {
	logFile, err :=os.Create(".logs")
	if err != nil {
		panic(err)
	}

	logBuffer := &framebuffer.Buffer{}
	framebuffer.InitBuffer(logBuffer)

	logger := logging.NewLogger(
		io.MultiWriter(logBuffer, logFile),
		logfmt.MainFormat,
		logging.LevelDebug,
		256,
	)

	msPerUpdate := float64(0.100)
	agentsAmount := uint64(10)
	respondersAmount := uint64(20)
	minChanceToCrash := float32(0.1)
	alertsCapacity := uint64(32)
	minChanceToHandle := float32(0.95)

	simulation.Init(
		msPerUpdate,
		agentsAmount,
		respondersAmount,
		minChanceToCrash,
		alertsCapacity,
		minChanceToHandle,
		logBuffer,
		logger,
	)
	ui.Init(simulation.CommandsSystem, logBuffer)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				ui.StopEventLoop()
				fmt.Println(err)
				fmt.Fprintln(os.Stderr, string(debug.Stack()))
				os.Exit(1)
			}
		}()

		simulation.RunEventLoop()
	}()
	ui.RunEventLoop()

	ui.DrawFinalTable()
}
