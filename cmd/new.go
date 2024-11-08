package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/tui/formInput"
	"github.com/spf13/cobra"
)

// newCmd rep
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("new called")
	},
}

// newTaskCmd represents the new command
var newTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Creates a new task with a form",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
`,
	Run: func(cmd *cobra.Command, args []string) {
		form := &formInput.NewTaskForm{}

		err := form.NewForm()
		if err != nil {
			log.Fatalf("Error creating form: %v", err)
		}

		if form.Submit {

			newTask := data.Task{
				Title:    form.TaskTitle,
				Priority: form.Priority,
				Status:   form.Status,
				Archived: form.Archived,
			}

			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}
			defer conn.Close()
			theTask := data.TaskTable{
				Task: newTask,
			}
			err = theTask.Create(conn)
			if err != nil {
				log.Fatalf("Error creating task: %v", err)
			}
		}
	},
}

// newProjectCmd represents the new command
var newProjectCmd = &cobra.Command{
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

			newTask := data.Area{
				Title:    form.AreaTitle,
				Status:   form.Status,
				Archived: form.Archived,
			}

			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}
			defer conn.Close()
			theArea := data.AreaTable{
				Area: newTask,
			}
			err = theArea.Create(conn)
			if err != nil {
				log.Fatalf("Error creating task: %v", err)
			}
		}
	},
}

// newTaskNoteCmd represents the new command
var newTaskNoteCmd = &cobra.Command{
	Use:   "note",
	Short: "Add a note to a specific task",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			log.Fatalf("Usage: note <task_id> <note_title> <note_path>")
		}

		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid task ID: %v", err)
		}

		noteTitle := args[1]
		notePath := args[2]

		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		note := data.Note{
			Title: noteTitle,
			Path:  notePath,
		}

		taskTable := data.TaskTable{}
		err = data.AddNoteToEntity(&taskTable, conn, taskID, note)
		if err != nil {
			log.Fatalf("Error adding note to task: %v", err)
		}

		fmt.Println("Note added to task successfully")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(newTaskCmd)
	newCmd.AddCommand(newProjectCmd)
	newCmd.AddCommand(newTaskNoteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
