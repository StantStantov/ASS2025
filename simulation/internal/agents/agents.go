package agents

import (
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"
	"math/rand"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type AgentSystem struct {
	AgentsIds               []AgentId
	MinChanceToCrash        float32

	MachineInfoBatchChannel chan []models.MachineInfo
	ArrayPool *pools.ArrayPool[AgentId]
	BatchPool *pools.ArrayPool[models.MachineInfo]

	Logger *logging.Logger
}

type AgentId = uint64

func NewAgentSystem(
	capacity uint64,
	minChanceToCrash float32,
	machineInfoBatchChannel chan []models.MachineInfo,
	arrayPool *pools.ArrayPool[AgentId],
	batchPool *pools.ArrayPool[models.MachineInfo],
	logger *logging.Logger,
) *AgentSystem {
	system := &AgentSystem{}

	system.AgentsIds = make([]AgentId, capacity)
	for i := range capacity {
		system.AgentsIds[i] = AgentId(i)
	}
	system.MachineInfoBatchChannel = machineInfoBatchChannel
	system.MinChanceToCrash = minChanceToCrash

	system.ArrayPool = arrayPool
	system.BatchPool = batchPool

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "agent_system")
	})

	return system
}

func ProcessAgentSystem(system *AgentSystem) {
	aliveServices := pools.GetArray(system.ArrayPool)
	deadServices := pools.GetArray(system.ArrayPool)
	machineInfoBatch := pools.GetArray(system.BatchPool)
	for _, id := range system.AgentsIds {
		currentChance := rand.Float32()

		crashed := currentChance >= system.MinChanceToCrash
		if crashed {
			deadServices = append(deadServices, id)

			machineInfo := models.MachineInfo{Id: id}
			machineInfoBatch = append(machineInfoBatch, machineInfo)
		} else {
			aliveServices = append(aliveServices, id)
		}
	}

	logging.GetThenSendInfo(
		system.Logger,
		"received new statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Integer(event, "agents.alive.amount", len(aliveServices))
			logfmt.Unsigneds(event, "agents.alive.ids", aliveServices...)
			logfmt.Integer(event, "agents.dead.amount", len(deadServices))
			logfmt.Unsigneds(event, "agents.dead.ids", deadServices...)
			return nil
		},
	)

	system.MachineInfoBatchChannel <- machineInfoBatch

	pools.PutArrays(system.ArrayPool, aliveServices, deadServices)
}
