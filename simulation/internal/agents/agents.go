package agents

import (
	"StantStantov/ASS/internal/buffer"
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"
	"math/rand"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type AgentSystem struct {
	AgentsIds        []AgentId
	MinChanceToCrash float32

	Buffer    *buffer.BufferSystem
	ArrayPool *pools.ArrayPool[AgentId]
	JobsPool  *pools.ArrayPool[models.Job]
	JobPool   *pools.JobPool

	Logger *logging.Logger
}

type AgentId = uint64

func NewAgentSystem(
	capacity uint64,
	minChanceToCrash float32,
	buffer *buffer.BufferSystem,
	arrayPool *pools.ArrayPool[AgentId],
	jobsPool *pools.ArrayPool[models.Job],
	jobPool *pools.JobPool,
	logger *logging.Logger,
) *AgentSystem {
	system := &AgentSystem{}

	system.AgentsIds = make([]AgentId, capacity)
	for i := range capacity {
		system.AgentsIds[i] = AgentId(i)
	}
	system.MinChanceToCrash = minChanceToCrash

	system.Buffer = buffer
	system.ArrayPool = arrayPool
	system.JobsPool = jobsPool
	system.JobPool = jobPool

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "agent_system")
	})

	return system
}

func ProcessAgentSystem(system *AgentSystem) {
	aliveServices := pools.GetArray(system.ArrayPool)
	deadServices := pools.GetArray(system.ArrayPool)
	jobsToSend := pools.GetArray(system.JobsPool)
	for _, id := range system.AgentsIds {
		currentChance := rand.Float32()

		crashed := currentChance >= system.MinChanceToCrash
		if crashed {
			deadServices = append(deadServices, id)

			machineInfo := models.MachineInfo{Id: id}
			job := pools.GetJob(system.JobPool)
			job.Id = id
			job.Alerts = append(job.Alerts, machineInfo)

			jobsToSend = append(jobsToSend, job)
		} else {
			aliveServices = append(aliveServices, id)
		}
	}

	logging.GetThenSendInfo(
		system.Logger,
		"received new statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "agents.alive.ids", aliveServices...)
			logfmt.Unsigneds(event, "agents.dead.ids", deadServices...)
			return nil
		},
	)

	buffer.EnqueueIntoBuffer(system.Buffer, jobsToSend...)

	pools.PutArrays(system.ArrayPool, aliveServices, deadServices)
	pools.PutArrays(system.JobsPool, jobsToSend)
}
