package formInput

import (
	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/tui"
	"github.com/charmbracelet/huh"
)

type NewAreaForm struct {
	AreaForm  *huh.Form
	AreaTitle string
	Status    data.StatusType
	Notes     []data.Note
	Archived  bool
	Submit    bool
}

func (n *NewAreaForm) NewAreaForm(theme huh.Theme) error {
	tui.ClearTerminalScreen()

	groups := []*huh.Group{
		huh.NewGroup(
			huh.NewInput().
				Title("What is the the name of the Project/Area?").
				Prompt(">").
				Value(&n.AreaTitle),

			huh.NewSelect[data.StatusType]().
				Title("Current Status?").
				Options(
					huh.NewOption("Not Started", data.StatusToDo),
					huh.NewOption("Planning", data.StatusPlanning),
					huh.NewOption("In Progress", data.StatusDoing),
					huh.NewOption("Done", data.StatusDone),
				).
				Value(&n.Status),
			huh.NewSelect[bool]().
				Title("Do you want to archive this project/area right away?").
				Options(
					huh.NewOption("No", false).Selected(true),
					huh.NewOption("Yes", true),
				).
				Value(&n.Archived),
			huh.NewConfirm().
				Title("Are you ready to save?").
				Affirmative("Yes").
				Negative("No").
				Value(&n.Submit),
		),
	}
	n.AreaForm = huh.NewForm(groups...)

	return n.AreaForm.WithTheme(&theme).Run()
}
