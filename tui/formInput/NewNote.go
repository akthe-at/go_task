package formInput

import (
	"os"
	"os/exec"

	"github.com/akthe-at/go_task/data"
	"github.com/charmbracelet/huh"
)

type NewNoteForm struct {
	NoteForm *huh.Form
	Title    string
	Path     string
	Type     data.NoteType
	Submit   bool
}

func (n *NewNoteForm) NewForm() error {
	// Clear the terminal before showing form
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	groups := []*huh.Group{
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
				Title("Are you ready to save your task?").
				Affirmative("Yes").
				Negative("No").
				Value(&n.Submit),
		),
	}
	n.NoteForm = huh.NewForm(groups...)

	return n.NoteForm.Run()
}
