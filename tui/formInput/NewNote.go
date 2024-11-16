package formInput

import (
	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/tui"
	"github.com/charmbracelet/huh"
)

type NewNoteForm struct {
	NoteForm *huh.Form
	Title    string
	Path     string
	Type     data.NoteType
	Submit   bool
}

func (n *NewNoteForm) NewNoteForm() error {
	tui.ClearTerminalScreen()

	noteGroups := []*huh.Group{
		huh.NewGroup(
			huh.NewInput().
				Title("What note do you want to add?").
				Prompt(">").
				Value(&n.Title),
			huh.NewInput().
				Title("What is the note path?").
				Prompt(">").
				Value(&n.Path),
			huh.NewSelect[data.NoteType]().
				Options(huh.NewOptions(data.TaskNoteType, data.AreaNoteType)...).
				Title("What type of note is this?").
				Description("Choose a type").
				Value(&n.Type),
			huh.NewConfirm().
				Title("Are you ready to save your note?").
				Affirmative("Yes").
				Negative("No").
				Value(&n.Submit),
		),
	}
	n.NoteForm = huh.NewForm(noteGroups...)

	return n.NoteForm.Run()
}
