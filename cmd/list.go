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
	"os"
	"strconv"
	"strings"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
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
		fmt.Println("list called without arguments")
	},
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "List a singular task",
	Long:  `This command is used for viewing information about a singular task.`,
	Run: func(cmd *cobra.Command, args []string) {
		var taskID int
		if len(args) == 0 {
			log.Fatal("You must supply a task ID!")
		}

		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Errorf("There was an error converting the task ID to an integer: %v", err)
		}

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Errorf("There was an error connecting to the database: %v", err)
		}

		queries := sqlc.New(conn)
		defer conn.Close()

		task, err := queries.ReadTask(ctx, int64(taskID))
		if err != nil {
			log.Errorf("There was an error reading the tasks from the database: %v", err)
		}

		table := styleTaskTable(task)
		fmt.Println(table)
	},
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List your tasks",
	Long: `This command is used for calling for a list of your tasks.

`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Errorf("There was an error connecting to the database: %v", err)
		}
		queries := sqlc.New(conn)
		defer conn.Close()

		tasks, err := queries.ReadTasks(ctx)
		if err != nil {
			log.Errorf("There was an error reading the tasks from the database: %v", err)
		}
		table := styleTasksTable(tasks)
		// FIXME: Does this need to be a special renderer?
		fmt.Println(table)
	},
}

var taskNotesCmd = &cobra.Command{
	Use:   "notes",
	Short: "List Task Notes",
	Long: `Use this command to get a list of associated task notes.
	Simply supply the ID listed next to the task in "list tasks"`,
	Run: func(cmd *cobra.Command, args []string) {
		var taskID int

		conn, err := db.ConnectDB()
		if err != nil {
			log.Errorf("There was an error connecting to the database: %v", err)
		}
		defer conn.Close()

		taskID, err = strconv.Atoi(args[0])
		if err != nil {
			log.Errorf("There was an error converting the task ID to an integer: %v", err)
		}

		note := data.Note{ID: taskID}
		notes, err := note.ReadAll(conn, data.TaskNoteType)
		if err != nil {
			log.Errorf("There was an error reading the areas/projects from the database: %v", err)
		}

		table := styleTaskNotesTable(notes)
		fmt.Println(table)
	},
}

var allNotesCmd = &cobra.Command{
	Use:   "notes",
	Short: "List All Notes",
	Long: `Use this command to get a list of all notes, regardless of note type.
	To use this command just type list notes`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := db.ConnectDB()
		if err != nil {
			log.Errorf("There was an error connecting to the database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)
		allNotes, err := queries.ReadAllNotes(ctx)
		if err != nil {
			log.Errorf("There was an error reading the notes the database: %v", err)
		}

		table := styleAllNotesTable(allNotes)
		fmt.Println(table)
	},
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List your projects/areas",
	Long: `This command is used for calling for a list of your areas/tasks.

`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Errorf("There was an error connecting to the database: %v", err)
		}

		queries := sqlc.New(conn)
		defer conn.Close()

		areas, err := queries.ReadAreas(ctx)
		if err != nil {
			log.Errorf("There was an error reading the areas/projects from the database: %v", err)
		}
		table := styleAreaTable(areas)
		fmt.Println(table)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(tasksCmd)
	listCmd.AddCommand(projectsCmd)
	listCmd.AddCommand(taskCmd)
	listCmd.AddCommand(allNotesCmd)
	taskCmd.AddCommand(taskNotesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func styleTasksTable(tasks []sqlc.ReadTasksRow) *table.Table {
	theme := tui.GetSelectedTheme()
	re := lipgloss.NewRenderer(os.Stdout)
	var (
		HeaderStyle  = re.NewStyle().Foreground(lipgloss.Color(theme.Secondary)).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(20)
		OddRowStyle  = CellStyle.Foreground(lipgloss.Color(theme.Secondary))
		EvenRowStyle = CellStyle.Foreground(lipgloss.Color(theme.Primary))
	)
	//

	var rows [][]string
	for _, task := range tasks {
		formattedDate := task.AgeInDays
		row := []string{
			fmt.Sprintf("%d", task.ID),
			task.Title,
			fmt.Sprintf("%f", formattedDate),
		}
		var formattedNotes string
		if task.NoteTitles != nil {
			note := task.NoteTitles.(string)
			notes := strings.Split(note, ",")
			if len(notes) > 2 {
				formattedNotes = strings.Join(notes[:2], ", ") + ", ..."
			} else {
				formattedNotes = note
			}
		}
		row = append(row, formattedNotes)

		rows = append(rows, row)
	}
	t := *table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Success))).
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

			if col == 4 {
				style = style.Width(45)
			}
			return style
		}).
		Headers("ID", "Task", "Age of Task", "Notes").
		Rows(rows...)
	return &t
}

func styleAreaTable(areas []sqlc.ReadAreasRow) *table.Table {
	theme := tui.GetSelectedTheme()
	re := lipgloss.NewRenderer(os.Stdout)
	var (
		HeaderStyle  = re.NewStyle().Foreground(lipgloss.Color(theme.Secondary)).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(20)
		OddRowStyle  = CellStyle.Foreground(lipgloss.Color(theme.Secondary))
		EvenRowStyle = CellStyle.Foreground(lipgloss.Color(theme.Primary))
	)

	var rows [][]string
	for _, area := range areas {
		row := []string{
			fmt.Sprintf("%d", area.ID),
			area.Title,
			fmt.Sprintf("%v", area.Status.String),
		}
		rows = append(rows, row)
	}
	t := *table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Success))).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				style = HeaderStyle
			case row%2 == 0:
				style = EvenRowStyle
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
		Headers("ID", "Name", "Status").
		Rows(rows...)
	return &t
}

func styleTaskNotesTable(notesList []data.NoteTable) *table.Table {
	theme := tui.GetSelectedTheme()
	re := lipgloss.NewRenderer(os.Stdout)
	var (
		HeaderStyle  = re.NewStyle().Foreground(lipgloss.Color(theme.Secondary)).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(20)
		OddRowStyle  = CellStyle.Foreground(lipgloss.Color(theme.Secondary))
		EvenRowStyle = CellStyle.Foreground(lipgloss.Color(theme.Primary))
	)

	var rows [][]string
	for _, note := range notesList {
		row := []string{
			fmt.Sprintf("%d", note.NoteID),
			note.NoteTitle,
			note.NotePath,
		}
		rows = append(rows, row)
	}
	t := *table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Success))).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				style = HeaderStyle
			case row%2 == 0:
				style = EvenRowStyle
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
		Headers("ID", "Title", "Path").
		Rows(rows...)
	return &t
}

func styleTaskTable(task sqlc.ReadTaskRow) *table.Table {
	theme := tui.GetSelectedTheme()
	re := lipgloss.NewRenderer(os.Stdout)
	var (
		HeaderStyle  = re.NewStyle().Foreground(lipgloss.Color(theme.Secondary)).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(20)
		OddRowStyle  = CellStyle.Foreground(lipgloss.Color(theme.Secondary))
		EvenRowStyle = CellStyle.Foreground(lipgloss.Color(theme.Primary))
	)

	row := []string{
		fmt.Sprintf("%d", task.TaskID),
		task.TaskTitle,
		fmt.Sprintf("%f", task.AgeInDays),
		fmt.Sprintf("%v", task.NoteTitle),
	}

	t := *table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Success))).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				style = HeaderStyle
			case row%2 == 0:
				style = EvenRowStyle
			default:
				style = OddRowStyle
			}

			if col == 0 {
				style = style.Width(5)
			}

			if col == 1 {
				style = style.Width(15)
			}

			if col == 4 {
				style = style.Width(45)
			}
			return style
		}).
		Headers("ID", "Task", "Age of Task", "Notes").
		Rows([][]string{row}...)
	return &t
}

func styleAllNotesTable(notes []sqlc.ReadAllNotesRow) *table.Table {
	theme := tui.GetSelectedTheme()
	re := lipgloss.NewRenderer(os.Stdout)
	var (
		HeaderStyle  = re.NewStyle().Foreground(lipgloss.Color(theme.Secondary)).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(20)
		OddRowStyle  = CellStyle.Foreground(lipgloss.Color(theme.Secondary))
		EvenRowStyle = CellStyle.Foreground(lipgloss.Color(theme.Primary))
	)

	var rows [][]string
	for _, note := range notes {
		row := []string{
			fmt.Sprintf("%d", note.ID),
			note.Title,
			note.Path,
			note.AreaOrTaskTitle,
			note.ParentType,
		}
		rows = append(rows, row)
	}
	t := *table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Success))).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				style = HeaderStyle
			case row%2 == 0:
				style = EvenRowStyle
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
		Headers("ID", "Title", "Path", "Area/Task", "Parent ID").
		Rows(rows...)
	return &t
}
