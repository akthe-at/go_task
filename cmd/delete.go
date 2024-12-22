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
	"context"
	"fmt"
	"strconv"

	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "The root command for deletion related commands.",
	Long:  `This command is the root command for all deletion related commands. Please use --help to see all of the available subcommands and any various flags or options.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The delete cmd invoked without any additional arguments. Please provide a subcommand.")
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
		var taskIDs []int64
		for _, task := range args {
			taskID, err := strconv.Atoi(task)
			if err != nil {
				log.Fatalf("Error converting task ID to integer: %v", err)
			}
			taskIDs = append(taskIDs, int64(taskID))
		}

		fmt.Println("delete cmd invoked for task(s):", taskIDs)
		if len(taskIDs) == 0 {
			log.Errorf("No task IDs provided - delete command requires at least one task ID")
			return
		}
		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)
		_, err = queries.DeleteTasks(ctx, taskIDs)
		if err != nil {
			log.Errorf("Error deleting task(s): %v", err)
		} else {
			log.Printf("Succesfully Deleted!")
		}
	},
}

var deleteAreaCmd = &cobra.Command{
	Use:   "area",
	Short: "Delete a area given a unique ID",
	Long: `
	You can delete a task by providing the unique task ID.
	You can use this command like this: go_task delete area <area_id>

	Additionally, you can provide multiple area IDs to delete multiple areas at once:
	"go_task delete area <area_id> <area_id> <area_id> ..."

	You can find the task ID by using the 'go_task list areas' command.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var areaIDs []int64
		for _, area := range args {
			areaID, err := strconv.Atoi(area)
			if err != nil {
				log.Errorf("Error converting task ID to integer: %v", err)
				return
			}
			areaIDs = append(areaIDs, int64(areaID))
		}

		fmt.Println("delete called for task(s):", areaIDs)
		if len(areaIDs) == 0 {
			log.Errorf("No area IDs provided - delete command requires at least one area ID")
			return
		}

		// ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		// queries := sqlc.New(conn)
		//
		// notesOption, err := cmd.Flags().GetString("notes")
		// if err != nil {
		// 	log.Fatalf("Error getting notes flag: %v", err)
		// }

		// switch notesOption {
		// case "all":
		// 	_, err := queries.DeleteMultipleAreas(ctx, projectIDs)
		// 	_, err = queries.DeleteNotesFromMultipleAreas(ctx, sql.NullInt64{int64(projectIDs), true})
		// case "some":
		// 	_, err := queries.DeleteMultipleAreas(ctx, projectIDs)
		// 	if err != nil {
		// 		log.Fatalf("Error deleting project(s): %v", err)
		// 	}
		// 	_, err = queries.DeleteNotesFromSingleArea(ctx, projectIDs)
		// case "one":
		// 	_, err = queries.DeleteSingleArea(ctx, projectIDs)
		// 	_, err = queries.DeleteNote(ctx, projectIDs)
		//
		// default:
		// 	_, err := queries.DeleteMultipleAreas(ctx, projectIDs)
		//
		// }
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteTaskCmd)
	deleteCmd.AddCommand(deleteAreaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	deleteAreaCmd.Flags().String("notes", "n", "Pass this flag to delete all, some, or none of the notes associated with the area.")
}
