package config

import (
	"os"

	"github.com/charmbracelet/log"
)

var UserSettings Config

type Config struct {
	Selected NoteSettings `toml:"selected"`
}

type NoteSettings struct {
	Editor      string `toml:"editor"`
	NotesPath   string `toml:"notes_path"`
	UseObsidian bool   `toml:"use_obsidian"`
	Theme       string `toml:"theme"`
}

// GetEditorConfig gets the editor from the config file
// If no editor is set in the config file, it falls back to $EDITOR
func GetEditorConfig() string {
	editor := UserSettings.Selected.Editor
	if editor == "" {
		log.Warnf("No editor set in config file, falling back to $EDITOR.")
		editor = os.Getenv("EDITOR")
	}
	return editor
}
