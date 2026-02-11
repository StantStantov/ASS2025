package buffer

func JobsTotal(system *BufferSystem) uint64 {
	return uint64(len(system.Values.Dense))
}

func AlertsTotal(system *BufferSystem) uint64 {
	total := uint64(0)
	values := system.Values.Dense
	for _, value := range values {
		job := value.Value
		alerts := job.Alerts

		total += uint64(len(alerts))
	}

	return total
}
