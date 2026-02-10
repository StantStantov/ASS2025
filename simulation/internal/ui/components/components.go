package components

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/ui/input"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var Buffer = &strings.Builder{}

type TickMsg tea.Msg

type MainMenu struct {
	Input *input.InputSystem
}

func DoTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (mainMenu MainMenu) Init() tea.Cmd {
	return DoTick()
}

func (mainMenu MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyPress := msg.String()
		input.ProcessKeyPress(mainMenu.Input, keyPress)
	case TickMsg:
		return mainMenu, DoTick()
	}

	return mainMenu, nil
}

func (mainMenu MainMenu) View() string {
	defer Buffer.Reset()

	if simulation.IsPaused {
		Buffer.WriteString("Paused\n")
	} else {
		Buffer.WriteString("Running\n")
	}
	Buffer.WriteByte('\n')

	fmt.Fprintf(Buffer, "All Agents:      %v\n", simulation.AgentsSystem.AgentsIds)
	Buffer.WriteByte('\n')

	fmt.Fprintf(Buffer, "All Responders:  %v\n", simulation.RespondersSystem.Responders)
	fmt.Fprintf(Buffer, "Free:            %v\n", simulation.RespondersSystem.FreeResponders.Dense)
	fmt.Fprintf(Buffer, "Busy:            %v\n", simulation.RespondersSystem.BusyResponders.Dense)

	return Buffer.String()
}
