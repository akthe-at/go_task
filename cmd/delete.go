/*
Copyright © 2024 Adam Kelly <arkelly111@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")
	},
}

var deleteTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Delete a task given a task ID",
	Long: `
	You can delete a task by providing the unique task ID.
	You can use this command like this: go_task delete task <task_id>
	
	You can also provide multiple task IDs to delete multiple tasks at once:
	"go_task delete task <task_id> <task_id> <task_id> ..."

	You can find the task ID by using the 'go_task list tasks' command.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var taskIDs []int
		for _, task := range args {
			taskID, err := strconv.Atoi(task)
			if err != nil {
				log.Errorf("Error converting task ID to integer: %v", err)
				return
			}
			taskIDs = append(taskIDs, taskID)
		}

		fmt.Println("delete called for task(s):", taskIDs)
		if len(taskIDs) == 0 {
			log.Errorf("No task IDs provided - delete command requires at least one task ID")
			return
		}

		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		task := data.Task{}
		err = task.DeleteMultiple(conn, taskIDs)
		if err != nil {
			log.Errorf("Error deleting task(s): %v", err)
			return
		}
	},
}

var deleteProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Delete a project/area given a unique ID",
	Long: `
	You can delete a task by providing the unique task ID.
	You can use this command like this: go_task delete project <project_id>

	Additionally, you can provide multiple project IDs to delete multiple projects at once:
	"go_task delete project <project_id> <project_id> <project_id> ..."

	You can find the task ID by using the 'go_task list projects' command.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var projectIDs []int
		for _, project := range args {
			projectID, err := strconv.Atoi(project)
			if err != nil {
				log.Errorf("Error converting task ID to integer: %v", err)
				return
			}
			projectIDs = append(projectIDs, projectID)
		}

		fmt.Println("delete called for task(s):", projectIDs)
		if len(projectIDs) == 0 {
			log.Errorf("No project IDs provided - delete command requires at least one project ID")
			return
		}

		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		projects := data.Area{}
		err = projects.DeleteMultiple(conn, projectIDs)
		if err != nil {
			log.Errorf("Error deleting project(s): %v", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteTaskCmd)
	deleteCmd.AddCommand(deleteProjectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
