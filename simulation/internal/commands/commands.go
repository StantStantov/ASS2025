package commands

import (
	"StantStantov/ASS/internal/state"
	"os"

	"github.com/StantStantov/rps/swamp/collections/ringbuffer"
)

type (
	CommandType uint8
	CommandFunc func()
)

const (
	QuitCommand CommandType = iota
	PauseCommand
)

var CommandsFuncs = []CommandFunc{
	func() {
		state.RestoreTerminal()

		os.Exit(0)
	},
	func() {
		state.IsPaused = !state.IsPaused
	},
}

type CommandsSystem struct {
	Commands *ringbuffer.RingBuffer[CommandType, CommandType]
}

func NewCommandsSystem() *CommandsSystem {
	system := &CommandsSystem{}

	system.Commands = ringbuffer.New[CommandType, CommandType](2)

	return system
}

func EnqueqeCommands(system *CommandsSystem, commands ...CommandType) {
	for _, command := range commands {
		ringbuffer.Enqueue(system.Commands, command)
	}
}

func ProcessCommandsSystem(system *CommandsSystem) {
	for ringbuffer.Length(system.Commands) != 0 {
		commandType, err := ringbuffer.Dequeue(system.Commands)
		if err != nil {
			continue
		}

		commandFunc := CommandsFuncs[commandType]

		commandFunc()
	}
}
