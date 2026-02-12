package ui

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/commands"
	"StantStantov/ASS/internal/ui/components"
	"StantStantov/ASS/internal/ui/input"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	Input *input.InputSystem

	Tea *tea.Program
)

func Init(commandsSystem *commands.CommandsSystem, logsBuffer *strings.Builder) {
	Input = input.NewInputSystem(commandsSystem)

	commands.RegisterCommand(commandsSystem, commands.QuitCommand, func() {
		Tea.Quit()
		Tea.Wait()
	})
	commands.RegisterCommand(commandsSystem, commands.PauseCommand, func() {
		simulation.IsPaused = !simulation.IsPaused
	})

	mainMenu := components.MainMenu{
		Input:      Input,
		Info:       components.InfoWindow{Buffer: &strings.Builder{}},
		Logs:       components.LogsWindow{Buffer: logsBuffer},
	}
	Tea = tea.NewProgram(
		mainMenu,
		tea.WithAltScreen(),
	)
}

func RunEventLoop() {
	Tea.Run()
}
