package formInput

import (
	"context"
	"path"
	"strconv"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui"
	"github.com/akthe-at/go_task/utils"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

type NewTaskForm struct {
	TaskTitle         string
	AreaAssignment    string
	Area              string
	ProjectAssignment string
	ProgProject       string
	TaskForm          *huh.Form
	Priority          data.PriorityType
	Status            data.StatusType
	Notes             []sqlc.Note
	Archived          bool
	Submit            bool
}

func (n *NewTaskForm) NewTaskForm(theme huh.Theme) error {
	tui.ClearTerminalScreen()
	options := fetchProgProjects()
	areaOptions := fetchAreas()

	taskGroups := []*huh.Group{
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
			huh.NewSelect[bool]().
				Title("Do you want to archive this task right away?").
				Options(
					huh.NewOption("No", false).Selected(true),
					huh.NewOption("Yes", true),
				).
				Value(&n.Archived),
			huh.NewSelect[string]().
				Title("Global or Project Specific Task?").
				Options(
					huh.NewOption("Global", "global"),
					huh.NewOption("Local", "local").Selected(true),
				).
				Value(&n.ProjectAssignment),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Value(&n.ProgProject).
				Title("Which Project Repo Did You Want to Assign This Task To?").
				Options(options...),
		).WithHideFunc(func() bool {
			return n.ProjectAssignment == "global"
		}),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Did you want to assign this task to a specific area?").
				Options(
					huh.NewOption("Yes", "yes").Selected(true),
					huh.NewOption("No", "no"),
				).
				Value(&n.AreaAssignment),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Value(&n.Area).
				Title("Which area did you want to assign this task to?").
				Options(areaOptions...),
		).WithHideFunc(func() bool {
			return n.AreaAssignment == "no"
		}),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you ready to save your task?").
				Affirmative("Yes").
				Negative("No").
				Value(&n.Submit),
		),
	}
	n.TaskForm = huh.NewForm(taskGroups...)

	return n.TaskForm.WithTheme(&theme).Run()
}

func fetchAreas() []huh.Option[string] {
	ctx := context.Background()
	conn, err := db.ConnectDB()
	if err != nil {
		log.Errorf("There was an error connecting to the database: %v", err)
	}

	queries := sqlc.New(conn)
	defer conn.Close()

	areas, err := queries.ReadAllAreas(ctx)
	if err != nil {
		return nil
	}
	var options []huh.Option[string]
	for _, area := range areas {
		options = append(options, huh.Option[string]{Value: strconv.FormatInt(area.ID, 10), Key: area.Title})
	}

	return options
}

func fetchProgProjects() []huh.Option[string] {
	projMap := make(map[string]string)
	existingProjects := make(map[string]bool)

	ctx := context.Background()
	conn, err := db.ConnectDB()
	if err != nil {
		log.Errorf("There was an error connecting to the database: %v", err)
	}

	queries := sqlc.New(conn)
	defer conn.Close()

	projects, err := queries.ReadAllProgProjects(ctx)
	if err != nil {
		return nil
	}
	for _, proj := range projects {
		projName := path.Base(proj)
		if !existingProjects[projName] {
			projMap[proj] = projName
			existingProjects[projName] = true
		}
	}

	ok, projectDir, err := utils.CheckIfProjDir()
	if err != nil {
		log.Fatalf("Error checking if project directory: %v", err)
	}
	if ok {
		projectDirBase := path.Base(projectDir)
		exists := false
		for _, proj := range projMap {
			if proj == projectDirBase {
				exists = true
				break
			}
		}
		if !exists {
			projMap[projectDir] = projectDirBase
		}
	}

	var options []huh.Option[string]
	for fullPath, baseName := range projMap {
		options = append(options, huh.Option[string]{Value: fullPath, Key: baseName})
	}

	return options
}
