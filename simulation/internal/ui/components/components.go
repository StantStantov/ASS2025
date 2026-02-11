package components

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"StantStantov/ASS/internal/ui/input"
	"fmt"
	"strings"

	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
	tea "github.com/charmbracelet/bubbletea"
)

var FrameBuffer = &strings.Builder{}

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

	status := "Paused"
	if !simulation.IsPaused {
		status = "Running"
	}

	fmt.Fprintf(FrameBuffer, "Simulation:\n")
	fmt.Fprintf(FrameBuffer, "Status:          %s\n", status)
	fmt.Fprintf(FrameBuffer, "Tick:            %v\n", simulation.TickCounter)
	fmt.Fprintf(FrameBuffer, "\n")

	fmt.Fprintf(FrameBuffer, "Agents:\n")
	fmt.Fprintf(FrameBuffer, "Ids:             %v\n", simulation.AgentsSystem.AgentsIds)
	fmt.Fprintf(FrameBuffer, "Silent:          %v\n", simulation.AgentsSystem.Silent)
	fmt.Fprintf(FrameBuffer, "Alarmed:         %v\n", simulation.AgentsSystem.Alarmed)
	fmt.Fprintf(FrameBuffer, "\n")

	jobsBufferedAmount := sparsemap.Length(simulation.Buffer.Values)
	jobsIds := make([]uint64, jobsBufferedAmount)
	jobs := make([]models.Job, jobsBufferedAmount)
	sparsemap.GetAllFromSparseMap(simulation.Buffer.Values, jobsIds, jobs)
	jobsAlertsAmounts := make([]int, len(jobs)) 
	for i := range jobsAlertsAmounts{
		job := jobs[i]
		jobsAlertsAmounts[i] = len(job.Alerts)
	}

	fmt.Fprintf(FrameBuffer, "Buffer:\n")
	fmt.Fprintf(FrameBuffer, "Ids:             %v\n", jobsIds)
	fmt.Fprintf(FrameBuffer, "Alerts:          %v\n", jobsAlertsAmounts)
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

	respondersFreeAmount := sparseset.Length(simulation.RespondersSystem.Free)
	respondersFree := make([]models.ResponderId, respondersFreeAmount)
	respondersFree = sparseset.GetAllFromSparseSet(simulation.RespondersSystem.Free, respondersFree)
	respondersBusyAmount := sparsemap.Length(simulation.RespondersSystem.Busy)
	respondersBusy := make([]models.ResponderId, respondersBusyAmount)
	respondersBusy = sparsemap.GetAllKeysFromSparseMap(simulation.RespondersSystem.Busy, respondersBusy)

	fmt.Fprintf(FrameBuffer, "Responders:\n")
	fmt.Fprintf(FrameBuffer, "Ids:             %v\n", simulation.RespondersSystem.Responders)
	fmt.Fprintf(FrameBuffer, "Free:            %v\n", respondersFree)
	fmt.Fprintf(FrameBuffer, "Busy:            %v\n", respondersBusy)
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

type frameMsg struct{}

func nextFrame() tea.Msg {
	return frameMsg{}
}
