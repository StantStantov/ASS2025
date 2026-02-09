package render

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/commands"
	"StantStantov/ASS/internal/ui/input"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Ui struct {
	Tea *tea.Program

	Input *input.InputSystem
}

func InitUi(ui *Ui, inputSystem *input.InputSystem, commandsSystem *commands.CommandsSystem) {
	mainMenu := MainMenu{Ui: ui}

	ui.Tea = tea.NewProgram(mainMenu)
	ui.Input = inputSystem

	commands.RegisterCommand(commandsSystem, commands.QuitCommand, func() {
		StopUi(ui)
		os.Exit(0)
	})
	commands.RegisterCommand(commandsSystem, commands.PauseCommand, func() {
		simulation.IsPaused = !simulation.IsPaused
	})
}

func RunUi(ui *Ui) {
	ui.Tea.Run()
}

func StopUi(ui *Ui) {
	ui.Tea.Quit()
	ui.Tea.Wait()
}

type MainMenu struct {
	Ui *Ui
}

func (mainMenu MainMenu) Init() tea.Cmd {
	return nil
}

func (mainMenu MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyPress := msg.String()
		input.ProcessKeyPress(mainMenu.Ui.Input, keyPress)
	}

	return mainMenu, nil
}

func (mainMenu MainMenu) View() string {
	if simulation.IsPaused {
		return "Paused"
	}
	return "Running"
}
