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

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/spf13/cobra"
)

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "This command is for interacting with the database.",
	Long:  `With these commands you can initialize/setup the database, or reset the database to clear it to an initial state.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

// dbCmd represents the db command
var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initial setup of DB",
	Long:  `This command will perform the initial setup of the sqlite database to hold the tasks and user data.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DB is being setup...")

		// Open a database connection
		conn, err := db.ConnectDB()
		if err != nil {
			fmt.Println("Error opening database:", err)
		}
		defer conn.Close()

		// FIXME: This path can't stay hard coded like this.
		if !db.IsSetup(conn) && !db.FileExists("../new_demo.db") {
			fmt.Println("Setting up the database...")

			err = db.SetupDB(conn)
			if err != nil {
				fmt.Println("Error setting up database:", err)
			}
		}

		// // Create and Insert a new Task
		test := data.Task{
			Title:    "do laundry",
			Priority: "high",
			Status:   "Pending",
			Archived: false,
		}
		test.Create(conn)

		err = test.Create(conn)
		if err != nil {
			fmt.Println("Error creating task:", err)
		}

		updated_task := data.Task{
			ID:             1,
			Title:          "do laundry again and again and again",
			UpdateArchived: false,
		}

		_, err = updated_task.Update(conn)
		if err != nil {
			fmt.Println("Error updating task:", err)
		}

		deleted_task := data.Task{ID: 1}
		err = deleted_task.Delete(conn)
		if err != nil {
			fmt.Println("Error deleting task:", err)
		}

		// Query Data
		results, err := conn.Query(`SELECT * FROM tasks`)
		if err != nil {
			fmt.Println("Error when querying data:", err)
		} else {
			// View Data / Close Connection
			defer results.Close()
			for results.Next() {
				var task data.Task
				err := results.Scan(&task.ID, &task.Title, &task.Priority, &task.Status, &task.Archived, &task.CreatedAt, &task.LastModified, &task.DueDate)
				if err != nil {
					fmt.Println("Error when scanning data:", err)
					continue
				}
				fmt.Printf("Task: %+v\n", task)
			}
			if err := results.Err(); err != nil {
				fmt.Println("Error when iterating results:", err)
			}
		}
		fmt.Println("Setup complete")
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the DB to its initial state.",
	Long: `This command will drop all of the tables, thus deleting all tasks, projects, and areas.
	The tables will be then recreated to their blank, default state.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Open a database connection
		conn, err := db.ConnectDB()
		if err != nil {
			fmt.Println("Error opening database:", err)
		}
		defer conn.Close()

		err = db.ResetDB(conn)
		if err != nil {
			fmt.Println("Error resetting database:", err)
		}
		fmt.Println("Database reset complete")
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbResetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
