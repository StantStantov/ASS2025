package models

type Job struct {
	Id     uint64
	Alerts []MachineInfo
}

type MachineInfo struct {
	Id uint64
}

func JobsToIds(jobs []Job, setBuffer []uint64) []uint64 {
	minLength := min(len(jobs), len(setBuffer))
	for i := range minLength {
		job := jobs[i]
		setBuffer[i] = job.Id
	}

	return setBuffer[:minLength]
}

func JobsPtrToIds(jobs []*Job, setBuffer []uint64) []uint64 {
	minLength := min(len(jobs), len(setBuffer))
	actualLength := minLength
	for i := range minLength {
		job := jobs[i]
		if job == nil {
			actualLength--

			continue
		}
		setBuffer[i] = job.Id
	}

	return setBuffer[:actualLength]
}
