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
	"database/sql"
	"fmt"
	"strconv"

	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var deleteNotes bool

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

		if len(args) < 1 {
			log.Fatalf("No task IDs provided - delete command requires at least one task ID")
		}

		for _, task := range args {
			taskID, err := strconv.Atoi(task)
			if err != nil {
				log.Fatalf("Error converting task ID to integer: %v", err)
			}
			taskIDs = append(taskIDs, int64(taskID))
		}
		fmt.Println("delete cmd invoked for task(s):", taskIDs)

		ctx := context.Background()
		conn, _, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)
		for _, taskID := range taskIDs {
			deletedTaskID, err := queries.DeleteTask(ctx, taskID)
			if err != nil {
				log.Fatalf("Error deleting task: %v", err)
			}
			noteRows, err := queries.ReadTaskNote(ctx, sql.NullInt64{Int64: deletedTaskID, Valid: true})
			if err != nil && err != sql.ErrNoRows {
				log.Fatalf("Error reading task notes: %v", err)
			}
			for _, row := range noteRows {
				_, err = queries.DeleteTaskBridgeNote(ctx, sqlc.DeleteTaskBridgeNoteParams{
					NoteID:       sql.NullInt64{Int64: row.ID, Valid: true},
					ParentTaskID: sql.NullInt64{Int64: deletedTaskID, Valid: true},
				})
				if err != nil {
					log.Fatalf("Error deleting bridge note: %v", err)
				}
			}
			_, err = queries.RecycleTaskID(ctx, taskID)
			if err != nil {
				log.Fatalf("Error recycling task ID: %v", err)
			}
		}
		fmt.Println("Succesfully Deleted!")
	},
}

var deleteAreaCmd = &cobra.Command{
	Use:   "area",
	Short: "Delete a area given a unique ID",
	Long: `
	You can delete an area by providing the unique area ID.
	You can use this command like this: go_task delete area <area_id>

	Additionally, you can provide multiple area IDs to delete multiple areas at once:
	"go_task delete area <area_id> <area_id> <area_id> ..."

	You can find the area ID by using the 'go_task list areas' command.

	If you want to delete the notes associated with the area, you can use the --notes flag.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var areaIDs []int64

		if len(args) < 1 {
			log.Fatalf("No area IDs provided - delete command requires at least one area ID")
		}
		for _, area := range args {
			areaID, err := strconv.Atoi(area)
			if err != nil {
				log.Fatalf("Error converting area ID to integer: %v", err)
			}
			areaIDs = append(areaIDs, int64(areaID))
		}

		fmt.Println("delete called for the following area(s):", areaIDs)

		ctx := context.Background()
		conn, _, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}

		tx, err := conn.Begin()
		if err != nil {
			log.Fatalf("addAreaNoteCmd: Error beginning transaction: %v", err)
		}
		defer conn.Close()
		defer tx.Rollback()
		queries := sqlc.New(conn)

		qtx := queries.WithTx(tx)

		for _, areaID := range areaIDs {
			_, err = queries.RecycleAreaID(ctx, areaID)
		}
		if err != nil {
			log.Fatalf("Error recycling area ID: %v", err)
		}

		_, err = qtx.DeleteMultipleAreas(ctx, areaIDs)
		if err != nil {
			log.Fatalf("Error deleting area(s): %v", err)
		}

		if deleteNotes {
			_, err = qtx.DeleteNotes(ctx, areaIDs)
			for i, noteID := range areaIDs {
				_, err := queries.DeleteNote(ctx, noteID)
				if err != nil {
					log.Fatalf("Error deleting task: %s", err)
				}
				bridgeNote := sqlc.DeleteAreaBridgeNoteParams{
					NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
					ParentAreaID: sql.NullInt64{Int64: areaIDs[i], Valid: true},
				}
				_, err = queries.DeleteAreaBridgeNote(ctx, bridgeNote)
				if err != nil {
					log.Fatalf("Error deleting bridge note: %v", err)
				}
				_, err = queries.RecycleNoteID(ctx, noteID)
				if err != nil {
					log.Fatalf("Error recycling note ID: %v", err)
				}
			}
			if err != nil {
				log.Fatalf("There was an error deleting the notes associated with the area(s): %v", err)
			}
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error committing transaction: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteTaskCmd)
	deleteCmd.AddCommand(deleteAreaCmd)

	deleteAreaCmd.PersistentFlags().BoolVar(&deleteNotes, "notes", false, "Pass this flag to delete the notes associated with the area that is to be deleted.")
}
