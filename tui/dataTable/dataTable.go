package datatable

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	data "github.com/akthe-at/go_task/data"
	db "github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/tui"
	"github.com/akthe-at/go_task/tui/formInput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyID        = "id"
	columnKeyTask      = "title"
	columnKeyPriority  = "priority"
	columnKeyStatus    = "status"
	columnKeyArchived  = "archived"
	columnKeyCreatedAt = "created_at"
	columnKeyDueDate   = "due_date"
)

var customBorder = table.Border{
	Top:    "─",
	Left:   "│",
	Right:  "│",
	Bottom: "─",

	TopRight:    "╮",
	TopLeft:     "╭",
	BottomRight: "╯",
	BottomLeft:  "╰",

	TopJunction:    "┬",
	LeftJunction:   "├",
	RightJunction:  "┤",
	BottomJunction: "┴",
	InnerJunction:  "┼",

	InnerDivider: "│",
}

type Model struct {
	tableModel    table.Model
	deleteMessage string
	archiveFilter bool
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.tableModel, cmd = m.tableModel.Update(msg)
	cmds = append(cmds, cmd)

	m.updateFooter()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)
		case "D", "dd":
			cmds = append(cmds, m.deleteTask())
		case "F":
			cmds = append(cmds, m.filterArchives())
		case "A":
			cmds = append(cmds, m.addTask())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) loadRowsFromDatabase() ([]table.Row, error) {
	conn, err := db.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	defer conn.Close()

	taskTable := data.TaskTable{}
	tasks, err := taskTable.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("error reading tasks: %w", err)
	}

	var rows []table.Row
	for _, task := range tasks {
		row := table.NewRow(table.RowData{
			columnKeyID:        fmt.Sprintf("%d", task.ID),
			columnKeyTask:      task.Title,
			columnKeyPriority:  task.Priority,
			columnKeyStatus:    task.Status,
			columnKeyArchived:  fmt.Sprintf("%t", task.Archived),
			columnKeyCreatedAt: task.CreatedAt.Format("2006-01-02 15:04:05"),
			columnKeyDueDate:   task.DueDate.Format("2006-01-02 15:04:05"),
		})
		rows = append(rows, row)
	}

	return rows, nil
}

func (m *Model) filterArchives() tea.Cmd {
	var filteredRows []table.Row
	// toggle m.archiveFilter from current status
	m.archiveFilter = !m.archiveFilter

	if m.archiveFilter {

		rows, err := m.loadRowsFromDatabase()
		if err != nil {
			log.Printf("Error loading rows from database: %s", err)
			return nil
		}

		for _, row := range rows {
			archived, ok := row.Data[columnKeyArchived]
			if !ok {
				log.Printf("Error getting archived status from row: %s", err)
				return nil
			}
			if archived == "false" {
				filteredRows = append(filteredRows, row)
			}
		}

		m.tableModel = m.tableModel.WithRows(filteredRows)

		// Update the footer
		m.updateFooter()

		return nil
	} else {
		rows, err := m.loadRowsFromDatabase()
		if err != nil {
			log.Printf("Error loading rows from database: %s", err)
			return nil
		}

		for _, row := range rows {
			archived, ok := row.Data[columnKeyArchived]
			if !ok {
				log.Printf("Error getting archived status from row: %s", err)
				return nil
			}
			if archived == "true" {
				filteredRows = append(filteredRows, row)
			}
		}

		m.tableModel = m.tableModel.WithRows(rows)

		// Update the footer
		m.updateFooter()

		return nil
	}
}

func (m *Model) addTask() tea.Cmd {
	form := &formInput.NewTaskForm{}

	err := form.NewForm()
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {

		newTask := data.Task{
			Title:    form.TaskTitle,
			Priority: form.Priority,
			Status:   form.Status,
			Archived: form.Archived,
		}

		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		theTask := data.TaskTable{
			Task: newTask,
		}
		err = theTask.Create(conn)
		if err != nil {
			log.Fatalf("Error creating task: %v", err)
		}
		// Requery the database and update the table model
		rows, err := m.loadRowsFromDatabase()
		if err != nil {
			log.Printf("Error loading rows from database: %s", err)
			return nil
		}
		m.tableModel = m.tableModel.WithRows(rows)

		// Update the footer
		m.updateFooter()
	}

	return nil
}

func (m *Model) deleteTask() tea.Cmd {
	selectedIDs := []string{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[columnKeyID].(string))
	}
	highlightedInfo := m.tableModel.HighlightedRow().Data[columnKeyID].(string)
	taskID, err := strconv.Atoi(highlightedInfo)
	if err != nil {
		log.Printf("Error converting ID to int: %s", err)
		return nil
	}

	conn, err := db.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %s", err)
		return nil
	}
	defer conn.Close()

	if len(selectedIDs) == 1 {
		task := data.Task{ID: taskID}
		err = task.Delete(conn)
		if err != nil {
			log.Printf("Error deleting task: %s", err)
			return nil
		}
		m.deleteMessage = fmt.Sprintf("You deleted this task:  IDs: %s", highlightedInfo)
	} else if len(selectedIDs) > 1 {
		deletedTasks := make([]string, len(selectedIDs))
		for idx, id := range selectedIDs {
			converted_id, err := strconv.Atoi(id)
			if err != nil {
				log.Printf("Error converting ID to int: %s", err)
			}
			task := data.Task{ID: converted_id}
			err = task.Delete(conn)
			if err != nil {
				log.Printf("Error deleting task: %s", err)
				return nil
			}
			deletedTasks[idx] = id
		}
		m.deleteMessage = fmt.Sprintf("You deleted these tasks:  IDs: %s", strings.Join(deletedTasks, ", "))
	}

	// Requery the database and update the table model
	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Printf("Error loading rows from database: %s", err)
		return nil
	}
	m.tableModel = m.tableModel.WithRows(rows)

	// Update the footer
	m.updateFooter()

	return nil
}

func (m Model) View() string {
	body := strings.Builder{}

	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Overlay)).Render("-Add new task by pressing 'A'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Overlay)).Render("-Filter Archived Tasks by pressing 'F'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Overlay)).Render("-Press left/right or page up/down to move between pages") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Overlay)).Render("-Press space/enter to select a row, q or ctrl+c to quit") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Overlay)).Render("-Press D/dd to delete row(s) after selecting them.") + "\n")

	selectedIDs := []string{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[columnKeyID].(string))
	}

	body.WriteString(
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(
				tui.Themes.RosePineMoon.Love)).
			Render(
				fmt.Sprintf("Selected IDs: %s", strings.Join(selectedIDs, ", "))) + "\n")

	if m.deleteMessage != "" {
		body.WriteString(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(
					tui.Themes.RosePineMoon.Love)).
				Render(m.deleteMessage) + "\n")
	}

	body.WriteString(m.tableModel.View())
	body.WriteString("\n")

	return body.String()
}

func NewModel() Model {
	columns := []table.Column{
		table.NewColumn(columnKeyID, "ID", 5).WithStyle(
			lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Pine)).
				Align(lipgloss.Center)),
		table.NewColumn(columnKeyTask, "Task", 10),
		table.NewColumn(columnKeyPriority, "Priority", 10),
		table.NewColumn(columnKeyStatus, "Status", 10),
		table.NewColumn(columnKeyArchived, "Archived", 10),
		table.NewColumn(columnKeyCreatedAt, "Created At", 20),
		table.NewColumn(columnKeyDueDate, "Due Date", 20),
	}

	model := Model{archiveFilter: true}
	var filteredRows []table.Row
	rows, err := model.loadRowsFromDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range rows {
		archived, ok := row.Data[columnKeyArchived]
		if !ok {
			log.Printf("Error getting archived status from row: %s", err)
		}
		if archived == "false" {
			filteredRows = append(filteredRows, row)
		}
	}

	keys := table.DefaultKeyMap()
	keys.RowDown.SetKeys("j", "down", "s")
	keys.RowUp.SetKeys("k", "up", "w")

	model.tableModel = table.New(columns).
		WithRows(filteredRows).
		HeaderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Foam)).Bold(true)).
		SelectableRows(true).
		Focused(true).
		Border(customBorder).
		WithKeyMap(keys).
		WithStaticFooter("Footer!").
		WithPageSize(10).
		WithSelectedText(" ", " 󰄲  ").
		WithBaseStyle(
			lipgloss.NewStyle().
				BorderForeground(lipgloss.Color(tui.Themes.RosePineMoon.Love)).
				Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Gold)).
				Align(lipgloss.Left),
		).
		SortByAsc(columnKeyID).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Love)),
			Data:  "<Missing Data>",
		})

	model.updateFooter()

	return model
}

func (m *Model) updateFooter() {
	highlightedRow := m.tableModel.HighlightedRow()
	rowID, ok := highlightedRow.Data[columnKeyID]
	if !ok {
		rowID = "No Rows Available"
	}

	footerText := fmt.Sprintf(
		"Pg. %d/%d - Currently looking at ID: %s",
		m.tableModel.CurrentPage(),
		m.tableModel.MaxPages(),
		rowID,
	)

	m.tableModel = m.tableModel.WithStaticFooter(footerText)
}

func RunModel(m *Model) {
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
