package state

import (
	"os"

	"golang.org/x/term"
)

var (
	TermFD       int
	TermOldState *term.State
	IsPaused     bool
)

func InitState() {
	TermFD = int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(TermFD)
	if err != nil {
		panic(err)
	}
	TermOldState = oldState
}

func RestoreTerminal() {
	term.Restore(TermFD, TermOldState)
}
