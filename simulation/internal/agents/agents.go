package agents

import (
	"StantStantov/ASS/internal/common/mempools"
	"StantStantov/ASS/internal/dispatchers"
	"StantStantov/ASS/internal/metrics"
	"StantStantov/ASS/internal/models"
	"math/rand"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type AgentSystem struct {
	AgentsIds        []AgentId
	MinChanceToCrash float32

	Dispatcher *dispatchers.DispatchSystem

	ArrayPool *mempools.ArrayPool[AgentId]
	JobsPool  *mempools.ArrayPool[models.Job]

	Metrics *metrics.MetricsSystem

	Logger *logging.Logger
}

type AgentId = uint64

func NewAgentSystem(
	capacity uint64,
	minChanceToCrash float32,
	dispatcher *dispatchers.DispatchSystem,
	arrayPool *mempools.ArrayPool[AgentId],
	jobsPool *mempools.ArrayPool[models.Job],
	metrics *metrics.MetricsSystem,
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

	system.Metrics = metrics

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "agent_system")
	})

	return system
}

func ProcessAgentSystem(system *AgentSystem) {
	silentServices := mempools.GetArray(system.ArrayPool)
	alarmingServices := mempools.GetArray(system.ArrayPool)
	defer mempools.PutArrays(system.ArrayPool, silentServices, alarmingServices)
	jobsToSave := mempools.GetArray(system.JobsPool)
	defer mempools.PutArrays(system.JobsPool, jobsToSave)

	for _, id := range system.AgentsIds {
		currentChance := rand.Float32()

		crashed := currentChance >= system.MinChanceToCrash
		if crashed {
			machineInfo := models.MachineInfo{Id: id}
			job := models.Job{
				Id:     id,
				Alerts: []models.MachineInfo{machineInfo},
			}

			jobsToSave = append(jobsToSave, job)
			alarmingServices = append(alarmingServices, id)
		} else {
			silentServices = append(silentServices, id)
		}
	}

	metrics.AddToMetric(system.Metrics, metrics.AgentsSilentCounter, uint64(len(silentServices)))
	metrics.AddToMetric(system.Metrics, metrics.AgentsAlarmingCounter, uint64(len(alarmingServices)))

	logging.GetThenSendInfo(
		system.Logger,
		"polled agents for new statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "agents.silent.ids", silentServices...)
			logfmt.Unsigneds(event, "agents.alarming.ids", alarmingServices...)

			return nil
		},
	)

	dispatchers.SaveAlerts(system.Dispatcher, jobsToSave...)
}
