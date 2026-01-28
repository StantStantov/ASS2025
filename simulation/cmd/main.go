package main

import (
	"StantStantov/ASS/internal/agents"
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


	logger := logging.NewLogger(os.Stdout, logfmt.MainFormat, logging.LevelDebug, 256)

	agentSystem := agents.NewAgentSystem(uint64(agentsAmount), 0.5, logger)

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

				lag -= msPerUpdate
			}
		}
	}
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
