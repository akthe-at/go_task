package config

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
