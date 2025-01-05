package formInput

import (
	"context"
	"fmt"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

type NewNoteForm struct {
	NoteForm *huh.Form
	Title    string
	Path     string
	Type     data.NoteType
	ParentID int
	Submit   bool
}

func (n *NewNoteForm) NewNoteForm(theme huh.Theme) error {
	tui.ClearTerminalScreen()
	taskOptions := fetchNoteParent(data.TaskNoteType)
	areaOptions := fetchNoteParent(data.AreaNoteType)

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
				Title("What type of note is this?").
				Description("Choose a type").
				Options(
					huh.NewOption("Task Note", data.TaskNoteType),
					huh.NewOption("Area Note", data.AreaNoteType),
				).
				Value(&n.Type),
		),
		huh.NewGroup(
			huh.NewSelect[int]().
				Value(&n.ParentID).
				Title("Which Task Did You Want to Assign This Note To?").
				Options(taskOptions...),
		).WithHideFunc(func() bool {
			return n.Type != data.TaskNoteType
		}),
		huh.NewGroup(
			huh.NewSelect[int]().
				Value(&n.ParentID).
				Title("Which Area Did You Want to Assign This Note To?").
				Options(areaOptions...),
		).WithHideFunc(func() bool {
			return n.Type != data.AreaNoteType
		}),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you ready to save your note?").
				Affirmative("Yes").
				Negative("No").
				Value(&n.Submit),
		),
	}
	n.NoteForm = huh.NewForm(noteGroups...)

	return n.NoteForm.WithTheme(&theme).Run()
}

type NewQuickNoteForm struct {
	NoteForm *huh.Form
	Title    string
	Path     string
	Submit   bool
}

func (n *NewQuickNoteForm) NewNoteForm(theme huh.Theme) error {
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
			huh.NewConfirm().
				Title("Are you ready to save your note?").
				Affirmative("Yes").
				Negative("No").
				Value(&n.Submit),
		),
	}
	n.NoteForm = huh.NewForm(noteGroups...)

	return n.NoteForm.WithTheme(&theme).Run()
}

func fetchNoteParent(selection data.NoteType) []huh.Option[int] {
	var options []huh.Option[int]
	ctx := context.Background()

	conn, _, err := db.ConnectDB()
	if err != nil {
		log.Errorf("There was an error connecting to the database: %v", err)
	}

	queries := sqlc.New(conn)
	defer conn.Close()

	switch selection {
	case data.TaskNoteType:
		tasks, err := queries.ReadAllTasks(ctx)
		if err != nil {
			return nil
		}
		for _, task := range tasks {
			options = append(options, huh.NewOption(fmt.Sprintf("%s - %d", task.Title, task.ID), int(task.ID)))
		}
	case data.AreaNoteType:
		areas, err := queries.ReadAreas(ctx)
		if err != nil {
			return nil
		}
		for _, area := range areas {
			options = append(options, huh.NewOption(area.Title, int(area.ID)))
		}
	}

	return options
}
