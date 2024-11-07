package cmd

import (
	"fmt"
	"log"

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

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(newTaskCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
