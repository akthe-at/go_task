package tui

import (
	"os"
	"os/exec"
)

func ClearTerminalScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
