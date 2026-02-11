package agents

import (
	"StantStantov/ASS/internal/simulation/dispatchers"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"math/rand"

	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/filters"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type AgentSystem struct {
	AgentsIds        []AgentId
	MinChanceToCrash float32

	Silent  []AgentId
	Alarmed []AgentId

	Dispatcher *dispatchers.DispatchSystem

	Metrics *metrics.MetricsSystem

	Logger *logging.Logger
}

type AgentId = uint64

func NewAgentSystem(
	capacity uint64,
	minChanceToCrash float32,
	dispatcher *dispatchers.DispatchSystem,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *AgentSystem {
	system := &AgentSystem{}

	system.AgentsIds = make([]AgentId, capacity)
	for i := range capacity {
		system.AgentsIds[i] = AgentId(i)
	}
	system.MinChanceToCrash = minChanceToCrash

	system.Silent = []AgentId{}
	system.Alarmed = []AgentId{}

	system.Dispatcher = dispatcher

	system.Metrics = metrics

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "agent_system")
	})

	return system
}

func ProcessAgentSystem(system *AgentSystem) {
	areAlarmed := make([]bool, len(system.AgentsIds))
	for i := range areAlarmed {
		currentChance := rand.Float32()
		alarmed := currentChance >= system.MinChanceToCrash
		areAlarmed[i] = alarmed
	}

	alarmedAmount, silentAmount := bools.CountBools[uint64, uint64](areAlarmed...)
	silentAgents := make([]AgentId, silentAmount)
	alarmedAgents := make([]AgentId, alarmedAmount)

	silentAgents, alarmedAgents = filters.SeparateByBools(silentAgents, alarmedAgents, system.AgentsIds, areAlarmed)

	jobs := make([]models.Job, len(alarmedAgents))
	for i, id := range alarmedAgents {
		machineInfo := models.MachineInfo{Id: id}
		job := models.Job{
			Id:     id,
			Alerts: []models.MachineInfo{machineInfo},
		}

		jobs[i] = job
	}

	system.Silent = silentAgents
	system.Alarmed = alarmedAgents

	dispatchers.SaveAlerts(system.Dispatcher, jobs...)

	metrics.AddToMetric(system.Metrics, metrics.AgentsSilentCounter, uint64(len(system.Silent)))
	metrics.AddToMetric(system.Metrics, metrics.AgentsAlarmingCounter, uint64(len(system.Alarmed)))

	logging.GetThenSendInfo(
		system.Logger,
		"polled agents for new statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "agents.silent.ids", system.Silent...)
			logfmt.Unsigneds(event, "agents.alarming.ids", system.Alarmed...)

			return nil
		},
	)
}
