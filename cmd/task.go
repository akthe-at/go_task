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

	"github.com/spf13/cobra"
)

var (
	Name   string
	taskID string
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "This is the parent command for all task related commands",
	Long: `This is the parent command for all task related cli commands and 
	has features like new, delete, list, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The new task command was called")
	},
}

var addCmd = &cobra.Command{
	Use:   "new",
	Short: "This is the command for creating new tasks via CLI",
	Long:  `This is the command for creating new tasks via CLI...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The new task command was called")
		if Name != "" {
			fmt.Println("A new task called", Name, "was created")
		} else {
			fmt.Println("A new task was created")
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "This is the command for deleting tasks via CLI",
	Long:  `This is the command for deleting tasks via CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The delete task command was called")
		if taskID != "" {
			fmt.Printf("You deleted the task with the id of %v", taskID)
		} else {
			fmt.Println("You deleted a task") // not really, they didn't provide an id...
		}
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(addCmd)
	taskCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	addCmd.Flags().StringVarP(&Name, "name", "n", "", "What do you like to be called?")
	deleteCmd.Flags().StringVarP(&taskID, "taskid", "t", "", "What task did you want to delete?")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// taskCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// taskCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
