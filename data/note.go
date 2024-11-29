package data

import (
	"bytes"
	"fmt"
	"os"

	"github.com/akthe-at/go_task/config"
	"gopkg.in/yaml.v3"
)

type NoteMetadata struct {
	Title   string   `yaml:"title"`
	id      string   `yaml:"id"`
	aliases []string `yaml:"aliases"`
	tags    []string `yaml:"tags"`
}

type NoteContent struct {
	Metadata NoteMetadata
	Body     string
}

func GenereateNoteID(title string) string {
	// Create note IDs in a Zettelkasten format with a timestamp and a suffix.
	// In this case a note with the title 'My new note'
	// will be given an ID that looks like '1657296016-my-new-note',
	// and therefore the file name '1657296016-my-new-note.md'
	var suffix string
	if title != "" {
	}
	return suffix
}

func GenerateMarkdownFile(note NoteContent, outputPath string) error {
	var content bytes.Buffer

	yamlFrontMatter, err := yaml.Marshal(note.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML front matter: %w", err)
	}

	content.WriteString("---\n")
	content.Write(yamlFrontMatter)
	content.WriteString("---\n")

	content.WriteString(note.Body)

	return os.WriteFile(outputPath, content.Bytes(), 0644)
}

func templateMarkdownNote(Title string, ID string, aliases []string, tags []string) error {
	note := NoteContent{
		Metadata: NoteMetadata{
			Title:   Title,
			id:      ID,
			aliases: aliases,
			tags:    tags,
		},
		Body: ``,
	}
	notesPath := config.UserSettings.NotesPath
	err := GenerateMarkdownFile(note, notesPath)
	if err != nil {
		return fmt.Errorf("failed to generate markdown file: %w", err)
	} else {
		fmt.Println("Markdown note generated successfully.")
	}

	return nil
}

// `
//     note_id_func = function(title)
//       -- Create note IDs in a Zettelkasten format with a timestamp and a suffix.
//       -- In this case a note with the title 'My new note' will be given an ID that looks
//       -- like '1657296016-my-new-note', and therefore the file name '1657296016-my-new-note.md'
//       local suffix = ""
//       if title ~= nil then
//         -- If title is given, transform it into valid file name.
//         suffix = title:gsub(" ", "-"):gsub("[^A-Za-z0-9-]", ""):lower()
//       else
//         -- If title is nil, just add 4 random uppercase letters to the suffix.
//         for _ = 1, 4 do
//           suffix = suffix .. string.char(math.random(65, 90))
//         end
//       end
//       return tostring(os.time()) .. "-" .. suffix
//     end,
