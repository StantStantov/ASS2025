package components

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/buffer"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/ui/input"
	"fmt"
	"strings"

	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
	tea "github.com/charmbracelet/bubbletea"
)

var FrameBuffer = &strings.Builder{}

type frameMsg struct{}

func nextFrame() tea.Msg {
	return frameMsg{}
}

type MainMenu struct {
	Input *input.InputSystem
}

func (mainMenu MainMenu) Init() tea.Cmd {
	return nextFrame
}

func (mainMenu MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyPress := msg.String()
		input.ProcessKeyPress(mainMenu.Input, keyPress)
	}

	return mainMenu, nextFrame
}

func (mainMenu MainMenu) View() string {
	defer FrameBuffer.Reset()

	lineWidth := 32

	if simulation.IsPaused {
		fmt.Fprintf(FrameBuffer, "Paused\n")
	} else {
		fmt.Fprintf(FrameBuffer, "Running\n")
	}
	fmt.Fprintf(FrameBuffer, "\n")

	fmt.Fprintf(FrameBuffer, "Agents:\n")
	fmt.Fprintf(FrameBuffer, "Ids:             %v\n", simulation.AgentsSystem.AgentsIds)
	fmt.Fprintf(FrameBuffer, "\n")

	jobsBufferedAmount := sparsemap.Length(simulation.Buffer.Values)
	jobsIds := make([]uint64, jobsBufferedAmount)
	jobsAlertsAmount := make([]int, jobsBufferedAmount)
	for i, entry := range simulation.Buffer.Values.Dense {
		jobsIds[i] = entry.Value.Id
		jobsAlertsAmount[i] = len(entry.Value.Alerts)
	}

	fmt.Fprintf(FrameBuffer, "Buffer:\n")
	fmt.Fprintf(FrameBuffer, "Ids Total:       %v\n", buffer.JobsTotal(simulation.Buffer))
	fmt.Fprintf(FrameBuffer, "Ids:             %v\n", jobsIds)
	fmt.Fprintf(FrameBuffer, "Alerts Total:    %v\n", buffer.AlertsTotal(simulation.Buffer))
	fmt.Fprintf(FrameBuffer, "Alerts:          %v\n", jobsAlertsAmount)
	fmt.Fprintf(FrameBuffer, "\n")

	jobsQueuedAmount := sparsemap.Length(simulation.Pool.Present)
	jobsQueuedIds := make([]uint64, jobsQueuedAmount)
	jobsQueuedIds = sparsemap.GetAllKeysFromSparseMap(simulation.Pool.Present, jobsQueuedIds)
	jobsQueuedLockedAmount := sparseset.Length(simulation.Pool.Locked)
	jobsQueuedLockedIds := make([]uint64, jobsQueuedLockedAmount)
	jobsQueuedLockedIds = sparseset.GetAllFromSparseSet(simulation.Pool.Locked, jobsQueuedLockedIds)

	fmt.Fprintf(FrameBuffer, "Pool:\n")
	fmt.Fprintf(FrameBuffer, "Ids:             %v\n", jobsQueuedIds)
	fmt.Fprintf(FrameBuffer, "Locked:          %v\n", jobsQueuedLockedIds)
	fmt.Fprintf(FrameBuffer, "\n")

	fmt.Fprintf(FrameBuffer, "Responders:\n")
	fmt.Fprintf(FrameBuffer, "Ids:             %v\n", simulation.RespondersSystem.Responders)
	fmt.Fprintf(FrameBuffer, "Free:            %v\n", simulation.RespondersSystem.FreeResponders.Dense)
	fmt.Fprintf(FrameBuffer, "Busy:            %v\n", simulation.RespondersSystem.BusyResponders.Dense)
	fmt.Fprintf(FrameBuffer, "\n")

	metricsAmount := len(simulation.MetricsSystem.Metrics)
	metricsToPrint := make([]metrics.Metric, metricsAmount)
	metricsToPrint = metrics.GetMetrics(simulation.MetricsSystem, metricsToPrint)

	fmt.Fprintf(FrameBuffer, "Metrics:\n")
	for _, metric := range metricsToPrint {
		spacesToPrint := lineWidth - len(metric.Name)
		fmt.Fprintf(FrameBuffer, "%s:%*v\n", metric.Name, spacesToPrint, metric.Value)
	}
	fmt.Fprintf(FrameBuffer, "\n")

	return FrameBuffer.String()
}
