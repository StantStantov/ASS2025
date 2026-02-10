package components

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/ui/input"
	"fmt"
	"strings"

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

	if simulation.IsPaused {
		fmt.Fprintf(FrameBuffer, "Paused\n")
	} else {
		fmt.Fprintf(FrameBuffer, "Running\n")
	}
	fmt.Fprintf(FrameBuffer, "\n")

	fmt.Fprintf(FrameBuffer, "All Agents:      %v\n", simulation.AgentsSystem.AgentsIds)
	fmt.Fprintf(FrameBuffer, "\n")

	fmt.Fprintf(FrameBuffer, "All Responders:  %v\n", simulation.RespondersSystem.Responders)
	fmt.Fprintf(FrameBuffer, "Free:            %v\n", simulation.RespondersSystem.FreeResponders.Dense)
	fmt.Fprintf(FrameBuffer, "Busy:            %v\n", simulation.RespondersSystem.BusyResponders.Dense)

	return FrameBuffer.String()
}
