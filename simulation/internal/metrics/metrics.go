package metrics

import (
	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type MetricsSystem struct {
	AgentsSilentTotalGauge   uint64
	AgentsAlarmingTotalGauge uint64

	AlertsBufferedTotalGauge uint64

	JobsBufferedTotalGauge uint64
	JobsPendingTotalGauge  uint64
	JobsUnlockedTotalGauge uint64
	JobsLockedTotalGauge   uint64

	RespondersFreeTotalGauge uint64
	RespondersBusyTotalGauge uint64

	Logger *logging.Logger
}

func NewMetricsSystem(
	logger *logging.Logger,
) *MetricsSystem {
	system := &MetricsSystem{}

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "dispatch_system")
	})

	return system
}

func ProcessMetricsSystem(system *MetricsSystem) {
	logging.GetThenSendInfo(
		system.Logger,
		"saved new metrics values",
		func(event *logging.Event, level logging.Level) error {
			logfmt.Unsigned(event, "agents.silent_total", system.AgentsSilentTotalGauge)
			logfmt.Unsigned(event, "agents.alarming_total", system.AgentsAlarmingTotalGauge)

			logfmt.Unsigned(event, "alerts.buffered_total", system.AlertsBufferedTotalGauge)

			logfmt.Unsigned(event, "jobs.buffered_total", system.JobsBufferedTotalGauge)
			logfmt.Unsigned(event, "jobs.pending_total", system.JobsPendingTotalGauge)
			logfmt.Unsigned(event, "jobs.unlocked_total", system.JobsUnlockedTotalGauge)
			logfmt.Unsigned(event, "jobs.locked_total", system.JobsLockedTotalGauge)

			logfmt.Unsigned(event, "responders.free_total", system.RespondersFreeTotalGauge)
			logfmt.Unsigned(event, "responders.busy_total", system.RespondersBusyTotalGauge)

			return nil
		},
	)
}

func SetAgentsSilentTotal(system *MetricsSystem, newValue uint64) {
	system.AgentsSilentTotalGauge = newValue
}

func SetAgentsAlarmingTotal(system *MetricsSystem, newValue uint64) {
	system.AgentsAlarmingTotalGauge = newValue
}

func SetAlertsBufferedTotal(system *MetricsSystem, newValue uint64) {
	system.AlertsBufferedTotalGauge = newValue
}

func SetJobsBufferedTotal(system *MetricsSystem, newValue uint64) {
	system.JobsBufferedTotalGauge = newValue
}

func SetJobsPendingTotal(system *MetricsSystem, newValue uint64) {
	system.JobsPendingTotalGauge = newValue
}

func SetJobsUnlockedTotal(system *MetricsSystem, newValue uint64) {
	system.JobsUnlockedTotalGauge = newValue
}

func SetJobsLockedTotal(system *MetricsSystem, newValue uint64) {
	system.JobsLockedTotalGauge = newValue
}

func SetRespondersFreeTotal(system *MetricsSystem, newValue uint64) {
	system.RespondersFreeTotalGauge = newValue
}

func SetRespondersBusyTotal(system *MetricsSystem, newValue uint64) {
	system.RespondersBusyTotalGauge = newValue
}
