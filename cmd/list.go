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
	"os"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/tui"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

const (
	purple    = lipgloss.Color("99")
	gray      = lipgloss.Color("245")
	lightGray = lipgloss.Color("241")
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list called")
	},
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List your tasks",
	Long: `This command is used for calling for a list of your tasks.

`,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := db.ConnectDB()
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		taskTable := data.TaskTable{}
		tasks, err := taskTable.ReadAll(conn)
		if err != nil {
			panic(err)
		}
		table := styleTable(tasks)
		fmt.Println(table)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(tasksCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func styleTable(tasks []data.Task) *table.Table {
	re := lipgloss.NewRenderer(os.Stdout)
	var (
		HeaderStyle  = re.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Pine)).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(20)
		OddRowStyle  = CellStyle.Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Pine))
		EvenRowStyle = CellStyle.Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Love))
	)
	//

	var rows [][]string
	for _, task := range tasks {
		formattedDate := task.DueDate.Format("January 2, 2006")
		row := []string{
			fmt.Sprintf("%d", task.ID),
			task.Title,
			formattedDate,
		}
		rows = append(rows, row)
	}
	t := *table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Gold))).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				style = HeaderStyle
			case row%2 == 0:
				style = EvenRowStyle
			case col == 3:
				style = style.Width(35)
			default:
				style = OddRowStyle
			}

			if col == 0 {
				style = style.Width(5)
			}

			if col == 1 {
				style = style.Width(15)
			}
			return style
		}).
		Headers("ID", "Task", "Due Date").
		Rows(rows...)
	return &t
}
