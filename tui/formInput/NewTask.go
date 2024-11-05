package formInput

import (
	"github.com/akthe-at/go_task/data"
	"github.com/charmbracelet/huh"
)

var (
	TaskTitle string
	Priority  data.PriorityType
	Status    data.StatusType
	Archived  bool
	Submit    bool
	TaskForm  *huh.Form
)

func NewTaskForm() huh.Form {
	TaskForm := huh.NewForm(
		huh.NewGroup(
			// Ask the user for a base burger and toppings.
			huh.NewInput().
				Title("What is the task?").
				Prompt(">").
				Value(&TaskTitle),

			// Let the user select multiple toppings.
			huh.NewSelect[data.PriorityType]().
				Title("Priority Level").
				Options(
					huh.NewOption("Low", data.PriorityTypeLow).Selected(true),
					huh.NewOption("Medium", data.PriorityTypeMedium),
					huh.NewOption("High", data.PriorityTypeHigh),
					huh.NewOption("Urgent", data.PriorityTypeUrgent),
				).
				Value(&Priority),

			// Option values in selects and multi selects can be any type you
			// want. We’ve been recording strings above, but here we’ll store
			// answers as integers. Note the generic "[int]" directive below.
			huh.NewSelect[data.StatusType]().
				Title("Current Status?").
				Options(
					huh.NewOption("Not Started", data.StatusToDo),
					huh.NewOption("Planning", data.StatusPlanning),
					huh.NewOption("In Progress", data.StatusDoing),
					huh.NewOption("Done", data.StatusDone),
				).
				Value(&Status),
		),

		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Do you want to archive this task right away?").
				Options(
					huh.NewOption("No", false).Selected(true),
					huh.NewOption("Yes", true),
				).
				Value(&Archived),
		),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you ready to save your task?").
				Affirmative("Yes").
				Negative("No").
				Value(&Submit),
		),
	)

	return *TaskForm
}
