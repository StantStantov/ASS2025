package models

type Job struct {
	Id     uint64
	Alerts []MachineInfo
}

type MachineInfo struct {
	Id uint64
}
