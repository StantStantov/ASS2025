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
	JobsPendingCounter
	JobsUnlockedCounter
	JobsLockedCounter
	RespondersFreeCounter
	RespondersBusyCounter
)

var MetricTypesNames = []string{
	"agents_silent_total",
	"agents_alarming_total",
	"jobs_pending_total",
	"jobs_unlocked_total",
	"jobs_locked_total",
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

func ProcessMetricsSystem(system *MetricsSystem) {
	metrics := make([]Metric, len(MetricTypesNames))
	metrics = GetMetrics(system, metrics)
	logging.GetThenSendInfo(
		system.Logger,
		"saved new metrics values",
		func(event *logging.Event, level logging.Level) error {
			for _, metric := range metrics {
				logfmt.Unsigned(event, "metrics."+metric.Name, metric.Value)
			}

			return nil
		},
	)
}

func AddToMetric(system *MetricsSystem, metric MetricType, value uint64) {
	atomicValue := system.Metrics[metric]
	atomic.AddUint64(atomicValue, value)
}
