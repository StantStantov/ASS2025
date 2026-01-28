package dispatchers

import (
	"StantStantov/ASS/internal/models"
	"StantStantov/ASS/internal/pools"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type DispatchSystem struct {
	MachineInfoBatchChannel chan []models.MachineInfo

	BatchPool *pools.ArrayPool[models.MachineInfo]

	Logger    *logging.Logger
}

func NewDispatchSystem(
	machineInfoBatchChannel chan []models.MachineInfo,
	batchPool *pools.ArrayPool[models.MachineInfo],
	logger *logging.Logger,
) *DispatchSystem {
	system := &DispatchSystem{}

	system.MachineInfoBatchChannel = machineInfoBatchChannel

	system.BatchPool = batchPool

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "dispatch_system")
	})

	return system
}

func ProcessDispatchSystem(system *DispatchSystem) {
	receivedAmount := len(system.MachineInfoBatchChannel)
	for range receivedAmount {
		machineInfoBatch := <-system.MachineInfoBatchChannel

		logging.GetThenSendDebug(
			system.Logger,
			"received machine info batch",
			func(event *logging.Event, level logging.Level) error {
				ids := make([]uint64, len(machineInfoBatch))
				for i, info := range machineInfoBatch {
					ids[i] = info.Id
				}

				logfmt.Integer(event, "machines.amount", len(machineInfoBatch))
				logfmt.Unsigneds(event, "machines.ids", ids...)
				return nil
			},
		)

		pools.PutArrays(system.BatchPool, machineInfoBatch)
	}
}
