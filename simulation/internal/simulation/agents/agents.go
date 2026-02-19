package agents

import (
	"StantStantov/ASS/internal/simulation/dispatchers"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"math/rand"

	"github.com/StantStantov/rps/swamp/behaivors/buffers"
	"github.com/StantStantov/rps/swamp/bools"
	"github.com/StantStantov/rps/swamp/filters"
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type AgentSystem struct {
	AgentsIds        []models.AgentId
	MinChanceToCrash float32

	Silent  []models.AgentId
	Alarmed []models.AgentId

	Dispatcher *dispatchers.DispatchSystem

	Metrics *metrics.MetricsSystem
	Logger  *logging.Logger
}

func NewAgentSystem(
	capacity uint64,
	minChanceToCrash float32,
	dispatcher *dispatchers.DispatchSystem,
	metrics *metrics.MetricsSystem,
	logger *logging.Logger,
) *AgentSystem {
	system := &AgentSystem{}

	system.AgentsIds = make([]models.AgentId, capacity)
	for i := range capacity {
		system.AgentsIds[i] = models.AgentId(i)
	}
	system.MinChanceToCrash = minChanceToCrash

	system.Silent = []models.AgentId{}
	system.Alarmed = []models.AgentId{}

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
	silentAgents := make([]models.AgentId, silentAmount)
	alarmedAgents := make([]models.AgentId, alarmedAmount)
	silentBuffer := &buffers.SetBuffer[models.AgentId, uint64]{Array: silentAgents}
	alarmedBuffer := &buffers.SetBuffer[models.AgentId, uint64]{Array: alarmedAgents}
	filters.SeparateByBools(silentBuffer, alarmedBuffer, system.AgentsIds, areAlarmed)

	alerts := make([][]models.MachineInfo, len(alarmedAgents))
	for i, id := range alarmedAgents {
		machineInfo := models.MachineInfo{Id: id}

		alerts[i] = []models.MachineInfo{machineInfo}
	}

	dispatchers.SaveAlerts(system.Dispatcher, alarmedAgents, alerts)

	system.Silent = silentAgents
	system.Alarmed = alarmedAgents

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
