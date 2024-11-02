package datatable

import (
	"fmt"
	"log"
	"os"

	data "github.com/akthe-at/go_task/data"
	db "github.com/akthe-at/go_task/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type TableModel struct {
	table *table.Table
}

func (t *TableModel) Init() tea.Cmd { return nil }

func (t *TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.table = t.table.Width(msg.Width)
		t.table = t.table.Height(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return t, tea.Quit
		}

	}
	return t, cmd
}

func (t *TableModel) View() string {
	return "\n" + t.table.String() + "\n"
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
			}
		}).
		Border(lipgloss.ThickBorder())

	return &TableModel{t}
}

func RunTableModel(m *TableModel) tea.Model {
	return m
}

func RunModel(m *TableModel) {
	if _, err := tea.NewProgram(RunTableModel(m), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
