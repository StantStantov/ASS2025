package metrics

import (
	"github.com/StantStantov/rps/swamp/atomic"

	"github.com/StantStantov/rps/swamp/logging"
	"github.com/StantStantov/rps/swamp/logging/logfmt"
)

type MetricType uint8

const (
	AgentsSilentCounter MetricType = iota
	AgentsAlarmingCounter
	JobsBufferedCounter
	JobsPendingCounter
	JobsSkippedCounter
	JobsLockedCounter
	JobsUnlockedCounter
	AlertsBufferedCounter
	RespondersFreeCounter
	RespondersBusyCounter
)

var MetricTypesNames = []string{
	"agents_silent_total",
	"agents_alarming_total",
	"jobs_added_to_buffer_total",
	"jobs_added_to_pool_total",
	"jobs_skipped_pool_total",
	"jobs_locked_total",
	"jobs_unlocked_total",
	"alerts_added_to_buffer_total",
	"responders_free_total",
	"responders_busy_total",
}

type MetricsSystem struct {
	Metrics []*atomic.Uint64

	Logger *logging.Logger
}

type Metric struct {
	Name  string
	Value uint64
}

func NewMetricsSystem(
	logger *logging.Logger,
) *MetricsSystem {
	system := &MetricsSystem{}

	system.Metrics = make([]*atomic.Uint64, len(MetricTypesNames))
	for i := range system.Metrics {
		system.Metrics[i] = atomic.NewUint64(0)
	}

	system.Logger = logging.NewChildLogger(logger, func(event *logging.Event) {
		logfmt.String(event, "from", "dispatch_system")
	})

	return system
}

func GetMetrics(system *MetricsSystem, setMetricBuffer []Metric) []Metric {
	minLength := min(len(setMetricBuffer), len(MetricTypesNames), len(system.Metrics))
	for i := range minLength {
		name := MetricTypesNames[i]
		atomicValue := system.Metrics[i]
		value := atomic.LoadUint64(atomicValue)

		metric := Metric{
			Name:  name,
			Value: value,
		}
		setMetricBuffer[i] = metric
	}

	return setMetricBuffer[:minLength]
}

func AddToMetric(system *MetricsSystem, metric MetricType, value uint64) {
	atomicValue := system.Metrics[metric]
	atomic.AddUint64(atomicValue, value)
}
