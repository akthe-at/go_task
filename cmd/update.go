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
		fmt.Println("update called")
	},
}

// TODO: This is going to require good helper text!!!
//
// updateTaskCmd represents the task update command
var updateTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := strconv.ParseInt(args[1], 10, 64)

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		queries := sqlc.New(conn)

		switch args[0] {
		case "title":
			queries.UpdateTaskTitle(ctx, sqlc.UpdateTaskTitleParams{
				Title: args[2],
				ID:    id,
			})

		case "priority":
			priority, err := mapToPriorityType(args[2])
			if err != nil {
				log.Fatalf("Invalid priority type: %v", err)
			}

			queries.UpdateTaskPriority(ctx, sqlc.UpdateTaskPriorityParams{
				Priority: sql.NullString{String: string(priority), Valid: true},
				ID:       id,
			})

		case "status":
			status, err := mapToStatusType(args[2])
			if err != nil {
				log.Fatalf("Invalid status type: %v", err)
			}

			queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
				Status: sql.NullString{String: string(status), Valid: true},
				ID:     id,
			})

		default:
			fmt.Println("default arg")
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
		id, _ := strconv.ParseInt(args[1], 10, 64)

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		queries := sqlc.New(conn)

		switch args[0] {
		case "title":
			_, err = queries.UpdateAreaTitle(ctx, sqlc.UpdateAreaTitleParams{
				Title: args[2],
				ID:    id,
			})
			if err != nil {
				log.Fatalf("Error updating area title: %v", err)
			}

		case "status":
			status, err := mapToStatusType(args[2])
			if err != nil {
				log.Fatalf("Invalid status type: %v", err)
			}

			_, err = queries.UpdateAreaStatus(ctx, sqlc.UpdateAreaStatusParams{
				Status: sql.NullString{String: string(status), Valid: true},
				ID:     id,
			})
			if err != nil {
				log.Fatalf("Error updating area status: %v", err)
			}

		case "archived":
			archiveState, err := strconv.ParseBool(args[2])
			if err != nil {
				log.Fatalf("Invalid archive state: %v", err)
			}

			_, err = queries.UpdateAreaArchived(ctx, sqlc.UpdateAreaArchivedParams{
				Archived: archiveState,
				ID:       id,
			})
			if err != nil {
				log.Fatalf("Error updating the archive status: %v", err)
			}

		default:
			fmt.Println("default arg")
			fmt.Println("test")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateTaskCmd)
	updateCmd.AddCommand(updateAreaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
