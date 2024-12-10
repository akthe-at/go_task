package data

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/akthe-at/go_task/config"
	"gopkg.in/yaml.v3"
)

type NoteMetadata struct {
	Title   string   `yaml:"Title"`
	ID      string   `yaml:"ID"`
	Aliases []string `yaml:"Aliases"`
	Tags    []string `yaml:"Tags"`
}

type NoteContent struct {
	Metadata NoteMetadata
	Body     string
}

func GenerateMarkdownFile(note NoteContent, outputPath string) (string, error) {
	var content bytes.Buffer

	yamlFrontMatter, err := yaml.Marshal(note.Metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML front matter: %w", err)
	}

	content.WriteString("---\n")
	content.Write(yamlFrontMatter)
	content.WriteString("---\n")

	content.WriteString(note.Body)

	outputPath = fmt.Sprintf("%s/%s.md", outputPath, note.Metadata.ID)

	return outputPath, os.WriteFile(outputPath, content.Bytes(), 0644)
}

func TemplateMarkdownNote(Title string, ID string, aliases []string, tags []string) (string, error) {
	note := NoteContent{
		Metadata: NoteMetadata{
			Title:   Title,
			ID:      ID,
			Aliases: aliases,
			Tags:    tags,
		},
		Body: `test`,
	}
	notesPath := config.UserSettings.Selected.NotesPath
	output, err := GenerateMarkdownFile(note, notesPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate markdown file: %w", err)
	} else {
		fmt.Println("Markdown note generated successfully.")
	}

	return output, nil
}

// GenereateNoteID
// Create note IDs in a Zettelkasten format with a timestamp and a suffix.
// In this case a note with the title 'My new note'
// will be given an ID that looks like '1657296016-my-new-note',
// and therefore the file name '1657296016-my-new-note.md'
func GenerateNoteID(title string) string {
	var suffix string
	if title != "" {
		suffix = strings.ToLower(strings.Replace(title, " ", "-", -1))
		re := regexp.MustCompile("[^A-Za-z0-9-]")
		suffix = re.ReplaceAllString(suffix, "$1")
	} else {
		for i := 0; i < 4; i++ {
			suffix = suffix + string(rune(65+rand.Intn(25)))
		}
	}
	return fmt.Sprintf("%d-%s", time.Now().Unix(), suffix)
}
