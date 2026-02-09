package input

import (
	"StantStantov/ASS/internal/commands"
	"StantStantov/ASS/internal/controls"
)

type InputSystem struct {
	CommandsSystem *commands.CommandsSystem
}

func NewInputSystem(
	commandsSystem *commands.CommandsSystem,
) *InputSystem {
	system := &InputSystem{}

	system.CommandsSystem = commandsSystem

	return system
}

func ProcessKeyPress(system *InputSystem, keyPress string) {
	keyName := controls.KeyName(keyPress)
	command, ok := controls.Keybindings[keyName]
	if !ok {
		return
	}

	commands.EnqueqeCommands(system.CommandsSystem, command)
}
