package formInput

import (
	"os"
	"os/exec"

	"github.com/akthe-at/go_task/data"
	"github.com/charmbracelet/huh"
)

type NewAreaForm struct {
	AreaForm  *huh.Form
	AreaTitle string
	Priority  data.PriorityType
	Status    data.StatusType
	Notes     []data.Note
	Archived  bool
	Submit    bool
}

func (n *NewAreaForm) NewAreaForm() error {
	// Clear the terminal before showing form
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	groups := []*huh.Group{
		huh.NewGroup(
			huh.NewInput().
				Title("What is the the name of the Project/Area?").
				Prompt(">").
				Value(&n.AreaTitle),

			huh.NewSelect[data.PriorityType]().
				Title("Priority Level").
				Options(
					huh.NewOption("Low", data.PriorityTypeLow).Selected(true),
					huh.NewOption("Medium", data.PriorityTypeMedium),
					huh.NewOption("High", data.PriorityTypeHigh),
					huh.NewOption("Urgent", data.PriorityTypeUrgent),
				).
				Value(&n.Priority),

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

	return n.AreaForm.Run()
}
