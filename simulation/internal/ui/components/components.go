package components

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/ui/input"

	tea "github.com/charmbracelet/bubbletea"
)

type MainMenu struct {
	Input *input.InputSystem
}

func (mainMenu MainMenu) Init() tea.Cmd {
	return nil
}

func (mainMenu MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyPress := msg.String()
		input.ProcessKeyPress(mainMenu.Input, keyPress)
	}

	return mainMenu, nil
}

func (mainMenu MainMenu) View() string {
	if simulation.IsPaused {
		return "Paused"
	}
	return "Running"
}
