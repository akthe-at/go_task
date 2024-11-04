package datatable

import (
	"fmt"
	"log"
	"os"

	data "github.com/akthe-at/go_task/data"
	db "github.com/akthe-at/go_task/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyID           = "id"
	columnKeyTask         = "title"
	columnKeyPriority     = "priority"
	columnKeyStatus       = "status"
	columnKeyArchived     = "archived"
	columnKeyCreatedAt    = "created_at"
	columnKeyLastModified = "last_modified"
	columnKeyDueDate      = "due_date"
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

	TopJunction:    "╥",
	LeftJunction:   "├",
	RightJunction:  "┤",
	BottomJunction: "╨",
	InnerJunction:  "╫",

	InnerDivider: "║",
}


	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)
		case "D", "dd":
			cmds = append(cmds, m.deleteTask())
		}
	}

	return m, tea.Batch(cmds...)
}

func NewTable() *TableModel {
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("#f6c177")).Bold(true)
	// selectedStyle := baseStyle.Foreground(lipgloss.Color("#9ccfd8")).Background(lipgloss.Color("#44415a"))

	headers := []string{"ID", "Title", "Description", "Priority", "Status", "Archived", "Created At", "Last Modified", "Due Date"}
	var rows [][]string

	// Open a database connection
	conn, err := db.ConnectDB()
	if err != nil {
		fmt.Println("Error opening database:", err)
	}
	defer conn.Close()

	taskTable := data.TaskTable{}
	tasks, err := taskTable.ReadAll(conn)
	if err != nil {
		log.Fatal("Error reading tasks:", err)
	}

	for _, task := range tasks {
		row := []string{
			fmt.Sprintf("%d", task.ID),
			task.Title,
			task.Description,
			task.Priority,
			task.Status,
			fmt.Sprintf("%t", task.Archived),
			task.CreatedAt.Format("2006-01-02 15:04:05"),
			task.LastModified.Format("2006-01-02 15:04:05"),
			task.DueDate.Format("2006-01-02 15:04:05"),
		}
		rows = append(rows, row)
	}

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("#908caa"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return headerStyle
			case row%2 == 0:
				return baseStyle.Foreground(lipgloss.Color("#3e8fb0"))
			default:
				return baseStyle.Foreground(lipgloss.Color("#eb6f92"))
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
		table.NewColumn(columnKeyLastModified, "Last Modified", 20),
		table.NewColumn(columnKeyDueDate, "Due Date", 20),
	}

	model := Model{}

	rows, err := model.loadRowsFromDatabase()
	if err != nil {
		log.Fatal(err)
	}

	keys := table.DefaultKeyMap()
	keys.RowDown.SetKeys("j", "down", "s")
	keys.RowUp.SetKeys("k", "up", "w")

	model.tableModel = table.New(columns).
		WithRows(rows).
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
			Data:  "<ない>",
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

func RunModel(m *TableModel) {
	if _, err := tea.NewProgram(RunTableModel(m), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
