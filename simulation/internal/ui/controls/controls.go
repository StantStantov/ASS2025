package controls

import (
	"StantStantov/ASS/internal/ui/commands"
)

type KeyName string

var (
	QuitKey KeyName = "q"
	PauseKey KeyName = " "
)

var Keybindings = map[KeyName]commands.CommandType{
	QuitKey: commands.QuitCommand,
	PauseKey: commands.PauseCommand,
}
