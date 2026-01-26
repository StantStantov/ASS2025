package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

func main() {
	logger := logging.NewLogger(os.Stdout, logfmt.MainFormat, logging.LevelDebug, 256)

	ctx, stopCtx := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stopCtx()

	msPerUpdate := 0.01

	previous := timeToFloat64(time.Now())
	lag := 0.0
	for {
		select {
		case <-ctx.Done():
			logging.GetThenSendEvent(
				logger,
				logging.LevelDebug,
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
				// update
				lag -= msPerUpdate
			}

			logging.GetThenSendEvent(
				logger,
				logging.LevelDebug,
				"tick passed",
				func(event *logging.Event, level logging.Level) error {
					logfmt.String(event, "tick.elapsed_from_previous", strconv.FormatFloat(elapsed, 'f', 3, 64))
					return nil
				},
			)
		}
	}
}

func timeToFloat64(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano() / 1e9)
}
