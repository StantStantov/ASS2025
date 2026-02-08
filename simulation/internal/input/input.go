package input

import (
	"StantStantov/ASS/internal/commands"
	"StantStantov/ASS/internal/controls"
	"os"
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

func ListenToInput(system *InputSystem) {
	for {
		keyPressed := make([]byte, 1)
		if _, err := os.Stdin.Read(keyPressed); err != nil {
			continue
		}

		keyName := controls.KeyName(keyPressed)
		command, ok := controls.Keybindings[keyName]
		if !ok {
			return
		}

		commands.EnqueqeCommands(system.CommandsSystem, command)
	}
}
