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
	"log"
	"strconv"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update root called without subcommand, please see help for more information")
	},
}

// updateTaskCmd represents the task update command
var updateTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Update a task's field",
	Long: `To update a task you must pass  the field you wish to to modify, followed by the id of the task, and the new value for that field. 
	For example, to update the title of a task with an id of 1 you would pass the following command:
	go_task update task title 1 "New Title"`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputID    = args[1]
			inputField = args[0]
			inputEdit  = args[2]
		)
		convertedID, err := strconv.ParseInt(inputID, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing id: %v", err)
		}

		ctx := context.Background()
		conn, _, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		queries := sqlc.New(conn)

		switch inputField {
		case "title":
			_, err = queries.UpdateTaskTitle(ctx, sqlc.UpdateTaskTitleParams{
				Title: inputEdit,
				ID:    convertedID,
			})
			if err != nil {
				log.Fatalf("Error updating task title: %v", err)
			}

		case "priority":
			priority, err := data.StringToPriorityType(inputEdit)
			if err != nil {
				log.Fatalf("Invalid priority type: %v", err)
			}

			_, err = queries.UpdateTaskPriority(ctx, sqlc.UpdateTaskPriorityParams{
				Priority: sql.NullString{String: string(priority), Valid: true},
				ID:       convertedID,
			})
			if err != nil {
				log.Fatalf("There was an error updating the task priority: %v", err)
			}

		case "status":
			status, err := data.StringToStatusType(inputEdit)
			if err != nil {
				log.Fatalf("Invalid status type: %v", err)
			}

			_, err = queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
				Status: sql.NullString{String: string(status), Valid: true},
				ID:     convertedID,
			})
			if err != nil {
				log.Fatalf("Error updating task status: %v", err)
			}

		case "area":
			areaID, err := strconv.ParseInt(inputEdit, 10, 64)
			if err != nil {
				log.Fatalf("Error parsing area id: %v", err)
			}
			_, err = queries.UpdateTaskArea(ctx, sqlc.UpdateTaskAreaParams{AreaID: sql.NullInt64{Int64: areaID, Valid: true}, ID: convertedID})
			if err != nil {
				log.Fatalf("Error updating task area: %v", err)
			}

		case "archived":
			archiveState, err := strconv.ParseBool(inputField)
			if err != nil {
				log.Fatalf("Invalid archive state: %v", err)
			}

			_, err = queries.UpdateTaskArchived(ctx, sqlc.UpdateTaskArchivedParams{
				Archived: archiveState,
				ID:       convertedID,
			})
			if err != nil {
				log.Fatalf("Error updating the archive status: %v", err)
			}

		default:
			fmt.Printf("Unknown field: %v", inputField)
		}
	},
}

// updateAreaCmd represents the project update command
var updateAreaCmd = &cobra.Command{
	Use:   "area",
	Short: "Update area details",
	Long: `You must pass the id for the area that you wish to update...followed by the field that you wish to
	update such as title, status, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputID    = args[1]
			inputField = args[0]
			inputEdit  = args[2]
		)
		convertedID, err := strconv.ParseInt(inputID, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing id: %v", err)
		}

		ctx := context.Background()
		conn, _, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		queries := sqlc.New(conn)

		switch inputField {
		case "title":
			_, err = queries.UpdateAreaTitle(ctx, sqlc.UpdateAreaTitleParams{
				Title: inputEdit,
				ID:    convertedID,
			})
			if err != nil {
				log.Fatalf("Error updating area title: %v", err)
			}

		case "status":
			status, err := data.StringToStatusType(inputField)
			if err != nil {
				log.Fatalf("Invalid status type: %v", err)
			}

			_, err = queries.UpdateAreaStatus(ctx, sqlc.UpdateAreaStatusParams{
				Status: sql.NullString{String: string(status), Valid: true},
				ID:     convertedID,
			})
			if err != nil {
				log.Fatalf("Error updating area status: %v", err)
			}

		case "archived":
			archiveState, err := strconv.ParseBool(inputField)
			if err != nil {
				log.Fatalf("Invalid archive state: %v", err)
			}

			_, err = queries.UpdateAreaArchived(ctx, sqlc.UpdateAreaArchivedParams{
				Archived: archiveState,
				ID:       convertedID,
			})
			if err != nil {
				log.Fatalf("Error updating the archive status: %v", err)
			}

		default:
			fmt.Printf("Unknown field: %v", inputField)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateTaskCmd)
	updateCmd.AddCommand(updateAreaCmd)
}
