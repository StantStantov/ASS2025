package agents

import (
	"math/rand"
	"sync"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type AgentSystem struct {
	AgentsIds []AgentId

	MinChanceToCrash float32

	Pool *arrayPool

	Logger *logging.Logger
}

type AgentId = uint64

func NewAgentSystem(
	capacity uint64,
	minChanceToCrash float32,
	logger *logging.Logger,
) *AgentSystem {
	system := &AgentSystem{}

	system.AgentsIds = make([]AgentId, capacity)
	for i := range capacity {
		system.AgentsIds[i] = AgentId(i)
	}
	system.MinChanceToCrash = minChanceToCrash

	system.Pool = newArrayPool(capacity)

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "agent_system")
	})

	return system
}

func ProcessAgentSystem(system *AgentSystem) {
	arrays := getArrays(system.Pool, 2)
	aliveServices := arrays[0]
	deadServices := arrays[1]
	for _, id := range system.AgentsIds {
		currentChance := rand.Float32()

		crashed := currentChance >= system.MinChanceToCrash
		if crashed {
			deadServices = append(deadServices, id)
		} else {
			aliveServices = append(aliveServices, id)
		}
	}

	logging.GetThenSendInfo(
		system.Logger,
		"received new statuses",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigneds(event, "agents.alive", aliveServices...)
			logfmt.Unsigneds(event, "agents.dead", deadServices...)
			return nil
		},
	)

	putArrays(system.Pool, arrays...)
}

type arrayPool struct {
	Arrays *sync.Pool
	MaxCapacity uint64
}

func newArrayPool(maxCapacity uint64) *arrayPool {
	arrayPool := &arrayPool{}

	arrayPool.Arrays = &sync.Pool{
		New: func() any {
			return newArray(maxCapacity)
		},
	}
	arrayPool.MaxCapacity = maxCapacity

	return arrayPool
}

func newArray(capacity uint64) []AgentId {
	return make([]AgentId, 0, capacity)
}

func getArrays(pool *arrayPool, amount uint64) [][]AgentId {
	arrays := make([][]AgentId, amount)
	for i := range amount {
		got := pool.Arrays.Get()
		array, ok := got.([]AgentId)
		if !ok {
			arrays[i] = nil
		} else {
			arrays[i] = array
		}
	}

	return arrays
}

func putArrays(pool *arrayPool, arrays ...[]AgentId) {
	for _, array := range arrays {
		array = array[:0:pool.MaxCapacity]
		pool.Arrays.Put(array)
	}
}
