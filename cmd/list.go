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
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type TableRow interface {
	ToRow() []string
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "This is the root cmd for listing tasks, areas, and notes.",
	Long:  `This is the root cmd for listing tasks, areas, and notes. You need to supply a subcommand to list tasks, areas, or notes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The list root cmd called without arguments, please provide a subcommand.")
	},
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "List a singular task",
	Long:  `This command is used for viewing information about a singular task. To view please provide the ID of the task.`,
	Run: func(cmd *cobra.Command, args []string) {
		var taskID int
		if len(args) == 0 {
			log.Fatal("You must supply a task ID!")
		}

		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("There was an error converting the task ID to an integer: %v", err)
		}

		ctx := context.Background()
		conn, _, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("There was an error connecting to the database: %v", err)
		}

		queries := sqlc.New(conn)
		defer conn.Close()

		task, err := queries.ReadTask(ctx, int64(taskID))
		if err != nil {
			log.Fatalf("There was an error reading the tasks from the database: %v", err)
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
		conn, _, err := db.ConnectDB()
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
		fmt.Println(table)
	},
}

var taskNotesCmd = &cobra.Command{
	Use:   "notes",
	Short: "List Task Notes",
	Long: `Use this command to get a list of associated task notes.
	Simply supply the ID listed next to the task in "list tasks"`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var taskID int

		conn, _, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("There was an error connecting to the database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)
		taskID, err = strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("There was an error converting the task ID to an integer: %v", err)
		}

		results, err := queries.ReadTaskNote(ctx,
			sql.NullInt64{Int64: int64(taskID)},
		)
		if err != nil {
			log.Fatalf("There was an error reading the task notes from the database: %v", err)
		}

		table := styleTaskNotesTable(results)
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

		conn, _, err := db.ConnectDB()
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
	Use:   "areas",
	Short: "List your areas",
	Long: `This command is used for calling for a list of your areas.

`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		conn, _, err := db.ConnectDB()
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
}

type TasksRowWrapper struct {
	sqlc.ReadTasksRow
}

func (t TasksRowWrapper) ToRow() []string {
	formattedPath := path.Base(t.Path.String)
	if formattedPath == "." {
		formattedPath = ""
	}

	formattedDate := t.AgeInDays
	var formattedNotes string
	if t.NoteTitles != nil {
		note := t.NoteTitles.(string)
		notes := strings.Split(note, ",")
		if len(notes) > 2 {
			formattedNotes = strings.Join(notes[:2], ", ") + ", ..."
		} else {
			formattedNotes = note
		}
	}

	return []string{
		fmt.Sprintf("%d", t.ID),
		t.Title,
		t.Priority.String,
		t.Status.String,
		fmt.Sprintf("%.2f Days", formattedDate),
		formattedNotes,
		formattedPath,
		t.ParentArea.String,
	}
}

type AreaRowWrapper struct {
	sqlc.ReadAreasRow
}

func (a AreaRowWrapper) ToRow() []string {
	return []string{
		fmt.Sprintf("%d", a.ID),
		a.Title,
		fmt.Sprintf("%v", a.Status.String),
	}
}

type TaskNoteRowWrapper struct {
	sqlc.ReadTaskNoteRow
}

func (n TaskNoteRowWrapper) ToRow() []string {
	return []string{
		fmt.Sprintf("%d", n.ID),
		n.Title,
		n.Path,
	}
}

type TaskRowWrapper struct {
	sqlc.ReadTaskRow
}

// ToRow converts the TaskRowWrapper to a slice of strings
func (t TaskRowWrapper) ToRow() []string {
	formattedPath := path.Base(t.ProgProj.String)
	if formattedPath == "." {
		formattedPath = ""
	}
	var formattedNotes string
	if t.NoteTitle != nil {
		note := t.NoteTitle.(string)
		notes := strings.Split(note, ",")
		if len(notes) > 2 {
			formattedNotes = strings.Join(notes[:2], ", ") + ", ..."
		} else {
			formattedNotes = note
		}
	}

	return []string{
		fmt.Sprintf("%d", t.TaskID),
		t.TaskTitle,
		t.Priority.String,
		t.Status.String,
		fmt.Sprintf("%.2f Days", t.AgeInDays),
		formattedNotes,
		formattedPath,
		t.ParentArea.String,
	}
}

type AllNotesRowWrapper struct {
	sqlc.ReadAllNotesRow
}

func (a AllNotesRowWrapper) ToRow() []string {
	return []string{
		fmt.Sprintf("%d", a.ID),
		a.Title,
		a.Path,
		a.AreaOrTaskTitle,
		a.ParentType,
	}
}

// styleTable is a helper function to style the table output
func styleTable(rows []TableRow, headers []string, colWidths map[int]int) *table.Table {
	theme := tui.GetSelectedTheme()
	re := lipgloss.NewRenderer(os.Stdout)
	var (
		HeaderStyle  = re.NewStyle().Foreground(lipgloss.Color(theme.Secondary)).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(10)
		OddRowStyle  = CellStyle.Foreground(lipgloss.Color(theme.Secondary))
		EvenRowStyle = CellStyle.Foreground(lipgloss.Color(theme.Primary))
	)

	var tableRows [][]string
	for _, row := range rows {
		tableRows = append(tableRows, row.ToRow())
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

			if width, ok := colWidths[col]; ok {
				style = style.Width(width)
			}

			return style
		}).
		Headers(headers...).
		// Widh(96).
		Rows(tableRows...)

	return &t
}

func styleTasksTable(tasks []sqlc.ReadTasksRow) *table.Table {
	var rows []TableRow
	for _, task := range tasks {
		rows = append(rows, TasksRowWrapper{task})
	}

	headers := []string{"ID", "Task", "Priority", "Status", "Task Age", "Notes", "Project", "Area"}
	colWidths := map[int]int{0: 2, 1: 15, 2: 10, 3: 10, 4: 10, 5: 15, 6: 10, 7: 10}
	return styleTable(rows, headers, colWidths)
}

func styleAreaTable(areas []sqlc.ReadAreasRow) *table.Table {
	var rows []TableRow
	for _, area := range areas {
		rows = append(rows, AreaRowWrapper{area})
	}

	headers := []string{"ID", "Name", "Status"}
	colWidths := map[int]int{0: 5, 1: 15}
	return styleTable(rows, headers, colWidths)
}

func styleTaskNotesTable(notesList []sqlc.ReadTaskNoteRow) *table.Table {
	var rows []TableRow
	for _, note := range notesList {
		rows = append(rows, TaskNoteRowWrapper{note})
	}
	headers := []string{"ID", "Title", "Path"}
	colWidths := map[int]int{0: 5, 1: 15}
	return styleTable(rows, headers, colWidths)
}

func styleTaskTable(task sqlc.ReadTaskRow) *table.Table {
	var rows []TableRow

	rows = append(rows, TaskRowWrapper{task})
	headers := []string{"ID", "Task", "Priority", "Status", "Task Age", "Notes", "Project", "Area"}
	colWidths := map[int]int{0: 2, 1: 15, 2: 10, 3: 10, 4: 10, 5: 15, 6: 10, 7: 10}
	return styleTable(rows, headers, colWidths)
}

func styleAllNotesTable(notes []sqlc.ReadAllNotesRow) *table.Table {
	var rows []TableRow
	for _, note := range notes {
		rows = append(rows, AllNotesRowWrapper{note})
	}
	headers := []string{"ID", "Title", "Path", "Parent Title", "Area/Task"}
	colWidths := map[int]int{0: 5, 1: 15, 2: 15, 3: 15, 4: 15}
	return styleTable(rows, headers, colWidths)
}
