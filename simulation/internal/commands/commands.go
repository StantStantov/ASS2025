package commands

import (
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

type CommandsSystem struct {
	Queue *ringbuffer.RingBuffer[CommandType, CommandType]
	Funcs []CommandFunc
}

func NewCommandsSystem() *CommandsSystem {
	system := &CommandsSystem{}

	system.Queue = ringbuffer.New[CommandType, CommandType](2)
	system.Funcs = make([]CommandFunc, 2)

	return system
}

func EnqueqeCommands(system *CommandsSystem, commands ...CommandType) {
	for _, command := range commands {
		ringbuffer.Enqueue(system.Queue, command)
	}
}

func ProcessCommandsSystem(system *CommandsSystem) {
	for ringbuffer.Length(system.Queue) != 0 {
		commandType, err := ringbuffer.Dequeue(system.Queue)
		if err != nil {
			continue
		}

		commandFunc := system.Funcs[commandType]

		commandFunc()
	}
}

func RegisterCommand(system *CommandsSystem, commandType CommandType, callback CommandFunc) {
	system.Funcs[commandType] = callback
}
