package config

var UserSettings NoteSettings

type NoteSettings struct {
	Editor       string `toml:"editor"`
	NotesPath    string `toml:"notes_path"`
	UserObsidian bool   `toml:"use_obsidian"`
}
