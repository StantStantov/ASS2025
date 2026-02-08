package input

import (
	"os"

	"github.com/StantStantov/rps/swamp/collections/ringbuffer"
	"golang.org/x/term"
)

type InputSystem struct {
	Inputs *ringbuffer.RingBuffer[[]byte, uint8]

	TermFD       int
	TermOldState *term.State
}

func NewInputSystem() *InputSystem {
	system := &InputSystem{}

	oldState, err := term.MakeRaw(system.TermFD)
	if err != nil {
		panic(err)
	}
	system.TermOldState = oldState

	system.TermFD = int(os.Stdin.Fd())
	system.Inputs = ringbuffer.New[[]byte, uint8](2)

	return system
}

func ListenToInput(system *InputSystem) {
	go func() {
		for {
			keyPressed := make([]byte, 1)
			if _, err := os.Stdin.Read(keyPressed); err != nil {
				continue
			}

			err := ringbuffer.Enqueue(system.Inputs, keyPressed)
			if err != nil {
				continue
			}
		}
	}()
}

func StopListening(system *InputSystem) {
	term.Restore(system.TermFD, system.TermOldState)
}

func ProcessInput(system *InputSystem) {
	keyPressed, err := ringbuffer.Dequeue(system.Inputs)
	if err != nil {
		return
	}

	switch string(keyPressed) {
	case "q":
		StopListening(system)

		os.Exit(0)
	default:
	}
}
