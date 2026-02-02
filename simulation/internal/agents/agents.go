package agents

import (
	"StantStantov/ASS/internal/dispatchers"
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"
	"math/rand"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type AgentSystem struct {
	AgentsIds        []AgentId
	MinChanceToCrash float32

	Dispatcher *dispatchers.DispatchSystem

	ArrayPool *pools.ArrayPool[AgentId]
	JobsPool  *pools.ArrayPool[*models.Job]
	JobPool   *pools.JobPool

	Logger *logging.Logger
}

type AgentId = uint64

func NewAgentSystem(
	capacity uint64,
	minChanceToCrash float32,
	dispatcher *dispatchers.DispatchSystem,
	arrayPool *pools.ArrayPool[AgentId],
	jobsPool *pools.ArrayPool[*models.Job],
	jobPool *pools.JobPool,
	logger *logging.Logger,
) *AgentSystem {
	system := &AgentSystem{}

	system.AgentsIds = make([]AgentId, capacity)
	for i := range capacity {
		system.AgentsIds[i] = AgentId(i)
	}
	system.MinChanceToCrash = minChanceToCrash

	system.Dispatcher = dispatcher

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
	jobsToSave := pools.GetArray(system.JobsPool)
	for _, id := range system.AgentsIds {
		currentChance := rand.Float32()

		crashed := currentChance >= system.MinChanceToCrash
		if crashed {
			deadServices = append(deadServices, id)

			machineInfo := models.MachineInfo{Id: id}
			job := &models.Job{Id: 0, Alerts: nil}
			job.Id = id
			job.Alerts = append(job.Alerts, machineInfo)

			jobsToSave = append(jobsToSave, job)
		} else {
			aliveServices = append(aliveServices, id)
		}
	}

	logging.GetThenSendInfo(
		system.Logger,
		"polled agents for new statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "agents.alive.ids", aliveServices...)
			logfmt.Unsigneds(event, "agents.dead.ids", deadServices...)
			return nil
		},
	)

	dispatchers.SaveAlerts(system.Dispatcher, jobsToSave...)

	pools.PutArrays(system.ArrayPool, aliveServices, deadServices)
	pools.PutArrays(system.JobsPool, jobsToSave)
}
