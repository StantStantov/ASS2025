package ptime

import "time"

func TimeNowInSeconds() float64 {
	timestamp := time.Now()

	return float64(timestamp.UnixNano() / 1e9)
}
