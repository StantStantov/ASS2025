package components

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/framebuffer"
	"StantStantov/ASS/internal/simulation/metrics"
	"StantStantov/ASS/internal/simulation/models"
	"StantStantov/ASS/internal/ui/input"
	"fmt"
	"strings"

	"github.com/StantStantov/rps/swamp/atomic"
	"github.com/StantStantov/rps/swamp/behaivors/buffers"
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
	"github.com/StantStantov/rps/swamp/collections/sparseset"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

type MainMenu struct {
	Input *input.InputSystem

	Info InfoWindow
	Logs LogsWindow
}

func (mainMenu MainMenu) Init() tea.Cmd {
	return nextFrame
}

func (mainMenu MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyPress := msg.String()
		input.ProcessKeyPress(mainMenu.Input, keyPress)
	case tea.WindowSizeMsg:
		borderWidth := style.GetHorizontalBorderSize()
		borderHeight := style.GetVerticalBorderSize()
		windowWidth := msg.Width - 2*borderWidth
		windowHeight := msg.Height - borderHeight

		infoTablesWidth := int(float32(windowWidth) * 0.25)
		infoTablesHeight := windowHeight
		mainMenu.Info = InfoWindow{Buffer: mainMenu.Info.Buffer, Model: viewport.New(infoTablesWidth, infoTablesHeight)}

		logsWidth := windowWidth - infoTablesWidth
		logsHeight := windowHeight
		mainMenu.Logs = LogsWindow{Buffer: mainMenu.Logs.Buffer, LogBuffer: mainMenu.Logs.LogBuffer, Model: viewport.New(logsWidth, logsHeight)}
	}

	return mainMenu, nextFrame
}

func (mainMenu MainMenu) View() string {
	infoWindow := mainMenu.Info.View()
	infoWindowStyled := style.Render(infoWindow)

	viewport := mainMenu.Logs.View()
	viewportStyled := style.Render(viewport)

	return lipgloss.JoinHorizontal(lipgloss.Left, infoWindowStyled, viewportStyled)
}

type LogsWindow struct {
	Buffer    *strings.Builder
	LogBuffer *framebuffer.Buffer

	viewport.Model
}

func (lw LogsWindow) Init() tea.Cmd {
	return nil
}

func (lw LogsWindow) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return lw, nil
}

func (lw LogsWindow) View() string {
	defer lw.Buffer.Reset()

	framebuffer.String(lw.LogBuffer, lw.Buffer)
	logsToRender := lw.Buffer.String()
	logsToRenderWrapped := lipgloss.NewStyle().Width(lw.Model.Width).Render(logsToRender)

	lw.Model.SetContent(logsToRenderWrapped)

	return lw.Model.View()
}

type InfoWindow struct {
	Buffer *strings.Builder
	viewport.Model
}

func (iw InfoWindow) Init() tea.Cmd {
	return nil
}

func (iw InfoWindow) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return iw, nil
}

func (iw InfoWindow) View() string {
	defer iw.Buffer.Reset()

	status := "Paused"
	if !simulation.IsPaused {
		status = "Running"
	}

	fmt.Fprintf(iw.Buffer, "Simulation:\n")
	fmt.Fprintf(iw.Buffer, "Status:    %s\n", status)
	fmt.Fprintf(iw.Buffer, "Tick:      %v\n", simulation.TickCounter)
	fmt.Fprintf(iw.Buffer, "\n")

	fmt.Fprintf(iw.Buffer, "Agents:\n")
	fmt.Fprintf(iw.Buffer, "Ids:       %v\n", simulation.AgentsSystem.AgentsIds)
	fmt.Fprintf(iw.Buffer, "Silent:    %v\n", simulation.AgentsSystem.Silent)
	fmt.Fprintf(iw.Buffer, "Alarmed:   %v\n", simulation.AgentsSystem.Alarmed)
	fmt.Fprintf(iw.Buffer, "\n")

	jobsBufferedAmount := sparsemap.Length(simulation.Buffer.Values)
	jobsIds := make([]uint64, jobsBufferedAmount)
	jobs := make([]buffers.SetBuffer[models.MachineInfo, uint64], jobsBufferedAmount)
	sparsemap.GetAllFromSparseMap(simulation.Buffer.Values, jobsIds, jobs)
	jobsAlertsAmounts := make([]uint64, len(jobs))
	for i := range jobsAlertsAmounts {
		job := jobs[i]
		jobsAlertsAmounts[i] = job.Length
	}

	fmt.Fprintf(iw.Buffer, "Buffer:\n")
	fmt.Fprintf(iw.Buffer, "Ids:       %v\n", jobsIds)
	fmt.Fprintf(iw.Buffer, "Alerts:    %v\n", jobsAlertsAmounts)
	fmt.Fprintf(iw.Buffer, "\n")

	jobsQueuedAmount := sparsemap.Length(simulation.Pool.Present)
	jobsQueuedIds := make([]uint64, jobsQueuedAmount)
	jobsQueuedIds = sparsemap.GetAllKeysFromSparseMap(simulation.Pool.Present, jobsQueuedIds)
	jobsQueuedLockedAmount := sparseset.Length(simulation.Pool.Locked)
	jobsQueuedLockedIds := make([]uint64, jobsQueuedLockedAmount)
	jobsQueuedLockedIds = sparseset.GetAllFromSparseSet(simulation.Pool.Locked, jobsQueuedLockedIds)

	fmt.Fprintf(iw.Buffer, "Pool:\n")
	fmt.Fprintf(iw.Buffer, "Ids:       %v\n", jobsQueuedIds)
	fmt.Fprintf(iw.Buffer, "Locked:    %v\n", jobsQueuedLockedIds)
	fmt.Fprintf(iw.Buffer, "\n")

	respondersFreeAmount := sparseset.Length(simulation.RespondersSystem.Free)
	respondersFree := make([]models.ResponderId, 0, respondersFreeAmount)
	respondersFreeEntries := simulation.RespondersSystem.Free.Dense
	for _, id := range respondersFreeEntries {
		respondersFree = append(respondersFree, id)
	}

	respondersBusyAmount := sparsemap.Length(simulation.RespondersSystem.Busy)
	respondersBusy := make([]models.ResponderId, 0, respondersBusyAmount)
	respondersBusyEntries := simulation.RespondersSystem.Busy.Dense
	for _, entry := range respondersBusyEntries {
		respondersBusy = append(respondersBusy, entry.Index)
	}

	fmt.Fprintf(iw.Buffer, "Responders:\n")
	fmt.Fprintf(iw.Buffer, "Ids:       %v\n", simulation.RespondersSystem.Responders)
	fmt.Fprintf(iw.Buffer, "Free:      %v\n", respondersFree)
	fmt.Fprintf(iw.Buffer, "Busy:      %v\n", respondersBusy)
	fmt.Fprintf(iw.Buffer, "\n")

	lineWidth := 36
	metricsAmount := len(simulation.MetricsSystem.Metrics)
	metricsToPrint := make([]metrics.Metric, metricsAmount)
	metricsToPrint = metrics.GetMetrics(simulation.MetricsSystem, metricsToPrint)

	fmt.Fprintf(iw.Buffer, "Metrics:\n")
	for _, metric := range metricsToPrint {
		spacesToPrint := lineWidth - len(metric.Name)
		fmt.Fprintf(iw.Buffer, "%s:%*v\n", metric.Name, spacesToPrint, metric.Value)
	}

	spacesToPrint := lineWidth - len("time_in_pool_seconds")
	timeAverage := float64(0)
	if simulation.Pool.SpentTimeInPool != 0 {
		timeAverage = simulation.Pool.SpentTimeInPool / float64(simulation.Pool.PoppedAmount)
	}
	fmt.Fprintf(iw.Buffer, "%s:%*.2f\n", "time_in_pool_seconds", spacesToPrint, timeAverage)

	spacesToPrint = lineWidth - len("rewrite_percentage")
	allAlertsAtomic := &simulation.MetricsSystem.Metrics[metrics.AlertsBufferedCounter]
	rewrittenAlertsAtomic := &simulation.MetricsSystem.Metrics[metrics.AlertsRewrittenCounter]
	allAlerts := atomic.LoadUint64(allAlertsAtomic)
	rewrittenAlerts := atomic.LoadUint64(rewrittenAlertsAtomic)
	rewritePercentage := float64(0)
	if rewrittenAlerts != 0 {
		rewritePercentage = float64(rewrittenAlerts) / float64(allAlerts)
	}
	fmt.Fprintf(iw.Buffer, "%s:%*.2f\n", "rewrite_percentage", spacesToPrint, rewritePercentage)

	spacesToPrint = lineWidth - len("duplicate_percentage")
	addedJobsAtomic := &simulation.MetricsSystem.Metrics[metrics.JobsPendingCounter]
	skippedJobsAtomic := &simulation.MetricsSystem.Metrics[metrics.JobsSkippedCounter]
	addedJobs := atomic.LoadUint64(addedJobsAtomic)
	skippedJobs := atomic.LoadUint64(skippedJobsAtomic)
	allJobs := addedJobs + skippedJobs
	duplicatePercentage := float64(0)
	if skippedJobs != 0 {
		duplicatePercentage = float64(skippedJobs) / float64(allJobs)
	}
	fmt.Fprintf(iw.Buffer, "%s:%*.2f\n", "duplicate_percentage", spacesToPrint, duplicatePercentage)

	spacesToPrint = lineWidth - len("load_percentage")
	freeRespsAtomic := &simulation.MetricsSystem.Metrics[metrics.RespondersFreeCounter]
	busyRespsAtomic := &simulation.MetricsSystem.Metrics[metrics.RespondersBusyCounter]
	freeResps := atomic.LoadUint64(freeRespsAtomic)
	busyResps := atomic.LoadUint64(busyRespsAtomic)
	allResps := freeResps + busyResps
	loadPercentage := float64(0)
	if skippedJobs != 0 {
		loadPercentage = float64(busyResps) / float64(allResps)
	}
	fmt.Fprintf(iw.Buffer, "%s:%*.2f\n", "load_percentage", spacesToPrint, loadPercentage)

	fmt.Fprintf(iw.Buffer, "\n")

	frame := iw.Buffer.String()

	iw.Model.SetContent(frame)

	return iw.Model.View()
}

type frameMsg struct{}

func nextFrame() tea.Msg {
	return frameMsg{}
}
