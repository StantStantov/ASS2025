package ui

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/commands"
	"StantStantov/ASS/internal/simulation/framebuffer"
	"StantStantov/ASS/internal/ui/components"
	"StantStantov/ASS/internal/ui/input"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var Tea *tea.Program

func Init(commandsSystem *commands.CommandsSystem, logsBuffer *framebuffer.Buffer) {
	input := input.NewInputSystem(commandsSystem)

	commands.RegisterCommand(commandsSystem, commands.QuitCommand, StopEventLoop)
	commands.RegisterCommand(commandsSystem, commands.PauseCommand, func() {
		simulation.IsPaused = !simulation.IsPaused
	})

	mainMenu := components.MainMenu{
		Input: input,
		Info:  components.InfoWindow{Buffer: &strings.Builder{}},
		Logs:  components.LogsWindow{Buffer: &strings.Builder{}, LogBuffer: logsBuffer},
	}
	Tea = tea.NewProgram(
		mainMenu,
		tea.WithAltScreen(),
	)
}

func RunEventLoop() {
	Tea.Run()
}

func StopEventLoop() {
	Tea.Quit()
	Tea.Wait()
}
