package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui/formInput"
	"github.com/spf13/cobra"
)

var (
	rawFlag  bool
	Archived bool
)

// addCmd Used for adding new tasks, projects, notes, etc.
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")
	},
}

// addTaskCmd represents the new command
var addTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Creates a new task with a form (or optionally raw string input)",
	Long: `
	Form Example: 'go_task add task' and follow the prompts.
	Raw Example: 'go_task add task <task_title> <task_priority> <task_status>'
	Valid task priorities: low, medium, high, urgent
	Valid task statuses: todo, planning, doing, done
	You can also optionally provided an archived status for the task using the --archived flag.

`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		if rawFlag {
			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}
			defer conn.Close()

			priority, err := mapToPriorityType(args[1])
			if err != nil {
				log.Fatalf("Invalid priority type: %v", err)
			}

			status, err := mapToStatusType(args[2])
			if err != nil {
				log.Fatalf("Invalid status type: %v", err)
			}

			theTask := sqlc.CreateTaskParams{
				Title:    args[0],
				Priority: sql.NullString{String: string(priority), Valid: true},
				Status:   sql.NullString{String: string(status), Valid: true},
				Archived: Archived,
			}

			queries := sqlc.New(conn)
			result, err := queries.CreateTask(ctx, theTask)
			if err != nil {
				log.Fatalf("Error creating task: %v", err)
			}
			fmt.Println("The Result was: ", result)

		} else {
			form := &formInput.NewTaskForm{}

			err := form.NewTaskForm()
			if err != nil {
				log.Fatalf("Error creating form: %v", err)
			}

			if form.Submit {

				conn, err := db.ConnectDB()
				if err != nil {
					log.Fatalf("Error connecting to database: %v", err)
				}
				defer conn.Close()

				queries := sqlc.New(conn)
				result, err := queries.CreateTask(ctx, sqlc.CreateTaskParams{
					Title:    form.TaskTitle,
					Priority: sql.NullString{String: string(form.Priority), Valid: true},
					Status:   sql.NullString{String: string(form.Status), Valid: true},
				})
				if err != nil {
					log.Fatalf("Error creating task: %v", err)
				}
				fmt.Println("The Result was: ", result)
			}
		}
	},
}

// addProjectCmd represents the new command
var addProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Creates a new project with a form",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
`,
	Run: func(cmd *cobra.Command, args []string) {
		form := &formInput.NewAreaForm{}

		err := form.NewAreaForm()
		if err != nil {
			log.Fatalf("Error creating form: %v", err)
		}

		if form.Submit {
			ctx := context.Background()
			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}
			defer conn.Close()

			queries := sqlc.New(conn)
			_, err = queries.CreateArea(ctx, sqlc.CreateAreaParams{
				Title:    form.AreaTitle,
				Status:   sql.NullString{String: string(form.Status), Valid: true},
				Archived: form.Archived,
			})
			if err != nil {
				log.Fatalf("AddProjectCmd: Error creating task: %v", err)
			}
		}
	},
}

// addTaskNoteCmd represents the new command
var addTaskNoteCmd = &cobra.Command{
	Use:   "note",
	Short: "Add a note to a specific task",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			log.Fatalf("You must provide at least 3 arguments! Usage: note <task_id> <note_title> <note_path>")
		}

		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid task ID: %v", err)
		}

		noteTitle := args[1]
		notePath := args[2]

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		tx, err := conn.Begin()
		if err != nil {
			log.Fatalf("addTaskNoteCmd: Error beginning transaction: %v", err)
		}
		defer tx.Rollback()
		defer conn.Close()

		queries := sqlc.New(conn)
		qtx := queries.WithTx(tx)
		noteID, err := qtx.CreateNote(ctx, sqlc.CreateNoteParams{
			Title: noteTitle,
			Path:  notePath,
		})
		if err != nil {
			fmt.Printf("addTaskNoteCmd: There was an error creating the note: %v", err)
		}

		_, err = qtx.CreateTaskBridgeNote(ctx, sqlc.CreateTaskBridgeNoteParams{
			NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
			ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
			ParentTaskID: sql.NullInt64{Int64: int64(taskID), Valid: true},
		})
		if err != nil {
			log.Fatalf("addTaskNoteCmd: Error creating task bridge note: %v", err)
		}
		tx.Commit()

		fmt.Println("Note added to task successfully")
	},
}

// addProjectNoteCmd represents the command for adding new Project/Area notes to an existing Project/Area
var addProjectNoteCmd = &cobra.Command{
	Use:   "note",
	Short: "Add a note to a specific area/project",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			log.Fatalf("You must provide at least 3 arguments! Usage: add project note <project_id> <note_title> <note_path>")
		}

		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid project ID: %v", err)
		}

		noteTitle := args[1]
		notePath := args[2]

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		tx, err := conn.Begin()
		if err != nil {
			log.Fatalf("addProjectNoteCmd: Error beginning transaction: %v", err)
		}
		defer tx.Rollback()
		defer conn.Close()

		queries := sqlc.New(conn)
		qtx := queries.WithTx(tx)
		noteID, err := qtx.CreateNote(ctx, sqlc.CreateNoteParams{
			Title: noteTitle,
			Path:  notePath,
		})
		if err != nil {
			fmt.Printf("addProjectNoteCmd: There was an error creating the note: %v", err)
		}

		_, err = qtx.CreateAreaBridgeNote(ctx, sqlc.CreateAreaBridgeNoteParams{
			NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
			ParentCat:    sql.NullInt64{Int64: int64(data.AreaNoteType), Valid: true},
			ParentAreaID: sql.NullInt64{Int64: int64(projectID), Valid: true},
		})
		if err != nil {
			log.Fatalf("addProjectNoteCmd: Error creating task bridge note: %v", err)
		}
		err = tx.Commit()
		if err != nil {
			fmt.Printf("addProjectNoteCmd: Error committing transaction: %v", err)
		} else {
			fmt.Println("Note added to Area/Project successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addTaskCmd)
	addCmd.AddCommand(addProjectCmd)
	addTaskCmd.AddCommand(addTaskNoteCmd)
	addProjectCmd.AddCommand(addProjectNoteCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.
	addCmd.PersistentFlags().BoolVarP(&rawFlag, "raw", "r", false, "Bypass using the form and use raw input instead")
	addCmd.PersistentFlags().BoolVar(&Archived, "archived", false, "Archive the task or project upon creation")
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func mapToPriorityType(input string) (data.PriorityType, error) {
	switch input {
	case "low":
		return data.PriorityTypeLow, nil
	case "medium":
		return data.PriorityTypeMedium, nil
	case "high":
		return data.PriorityTypeHigh, nil
	case "urgent":
		return data.PriorityTypeUrgent, nil
	default:
		return "", errors.New("invalid priority type")
	}
}

func mapToStatusType(input string) (data.StatusType, error) {
	switch input {
	case "todo":
		return data.StatusToDo, nil
	case "planning":
		return data.StatusPlanning, nil
	case "doing":
		return data.StatusDoing, nil
	case "done":
		return data.StatusDone, nil
	default:
		return "", errors.New("invalid status type")
	}
}
