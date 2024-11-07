package formInput

import (
	"os"
	"os/exec"

	"github.com/akthe-at/go_task/data"
	"github.com/charmbracelet/huh"
)

type NewTaskForm struct {
	TaskForm  *huh.Form
	TaskTitle string
	Priority  data.PriorityType
	Status    data.StatusType
	Notes     []data.Note
	Archived  bool
	Submit    bool
}

func (n *NewTaskForm) NewForm() error {
	// Clear the terminal before showing form

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	groups := []*huh.Group{
		huh.NewGroup(
			huh.NewInput().
				Title("What is the task?").
				Prompt(">").
				Value(&n.TaskTitle),

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
		),
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Do you want to archive this task right away?").
				Options(
					huh.NewOption("No", false).Selected(true),
					huh.NewOption("Yes", true),
				).
				Value(&n.Archived),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you ready to save your task?").
				Affirmative("Yes").
				Negative("No").
				Value(&n.Submit),
		),
	}
	n.TaskForm = huh.NewForm(groups...)

	return n.TaskForm.Run()
}
