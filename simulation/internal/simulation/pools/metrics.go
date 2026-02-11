package pools

func JobsPendingTotal(system *PoolSystem) uint64 {
	return system.Queue.Length
}

func JobsUnlockedTotal(system *PoolSystem) uint64 {
	return JobsPendingTotal(system) - JobsLockedTotal(system)
}

func JobsLockedTotal(system *PoolSystem) uint64 {
	return uint64(len(system.Locked.Dense))
}
