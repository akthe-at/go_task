package datatable

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	data "github.com/akthe-at/go_task/data"
	db "github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui"
	"github.com/akthe-at/go_task/tui/formInput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"

	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyID         = "id"
	columnKeyTask       = "title"
	columnKeyPriority   = "priority"
	columnKeyStatus     = "status"
	columnKeyArchived   = "archived"
	columnKeyCreatedAt  = "created_at"
	columnKeyTaskAge    = "age_in_days"
	columnKeyDueDate    = "due_date"
	columnKeyNotes      = "notes"
	minWidth            = 150
	minHeight           = 5
	fixedVerticalMargin = 60
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

// This is the task table "screen" model
type TaskModel struct {
	tableModel       table.Model
	totalWidth       int
	totalHeight      int
	horizontalMargin int
	verticalMargin   int
	deleteMessage    string
	archiveFilter    bool
}

// Init initializes the model (can use this to run commands upon model initialization)
func (m *TaskModel) Init() tea.Cmd { return nil }

func (m *TaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "F":
			cmds = append(cmds, m.filterArchives())
		case "T":
			cmds = append(cmds, m.addTask())
		case "left":
			if m.calculateWidth() > minWidth {
				m.horizontalMargin++
				m.recalculateTable()
			}
		case "right":
			if m.horizontalMargin > 0 {
				m.horizontalMargin--
				m.recalculateTable()
			}
		case "up":
			if m.calculateHeight() > minHeight {
				m.verticalMargin++
				m.recalculateTable()
			}
		case "down":
			if m.verticalMargin > 0 {
				m.verticalMargin--
				m.recalculateTable()
			}
		}
	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.totalHeight = msg.Height

		m.recalculateTable()
	case SwitchToPreviousViewMsg:
		m.recalculateTable()
	}

	return m, tea.Batch(cmds...)
}

func (m *TaskModel) recalculateTable() {
	m.tableModel = m.tableModel.
		WithTargetWidth(m.calculateWidth()).
		WithMinimumHeight(m.calculateHeight())
}

func (m TaskModel) calculateWidth() int {
	return m.totalWidth - m.horizontalMargin
}

func (m TaskModel) calculateHeight() int {
	return m.totalHeight - m.verticalMargin - fixedVerticalMargin
}

func (m *TaskModel) loadRowsFromDatabase() ([]table.Row, error) {
	ctx := context.Background()
	conn, err := db.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	queries := sqlc.New(conn)
	defer conn.Close()

	tasks, err := queries.ReadTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading tasks: %w", err)
	}

	var rows []table.Row
	for _, task := range tasks {
		row := table.NewRow(table.RowData{
			columnKeyID:       fmt.Sprintf("%d", task.ID),
			columnKeyTask:     task.Title,
			columnKeyPriority: task.Priority.String,
			columnKeyStatus:   task.Status.String,
			columnKeyArchived: fmt.Sprintf("%t", task.Archived.Bool),
			columnKeyTaskAge:  task.AgeInDays,
			columnKeyNotes:    task.NoteTitles,
		})
		rows = append(rows, row)
	}
	return rows, nil
}

func (m *TaskModel) filterArchives() tea.Cmd {
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

func (m *TaskModel) addNote() tea.Cmd {
	form := &formInput.NewNoteForm{}
	err := form.NewNoteForm()
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		selectedIDs := []string{}

		for _, row := range m.tableModel.SelectedRows() {
			selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(string))
		}
		highlightedInfo := fmt.Sprintf("%v", m.tableModel.HighlightedRow().Data[NoteColumnKeyID])
		taskID, err := strconv.Atoi(highlightedInfo)
		if err != nil {
			log.Printf("Error converting ID to int: %s", err)
			return nil
		}

		newNote := sqlc.CreateNoteParams{
			Title: form.Title,
			Path:  form.Path,
		}

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}

		queries := sqlc.New(conn)
		defer conn.Close()
		noteID, err := queries.CreateNote(ctx, newNote)
		if err != nil {
			log.Fatalf("Error creating note: %v", err)
		}
		id, err := queries.CreateTaskBridgeNote(ctx, sqlc.CreateTaskBridgeNoteParams{
			NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
			ParentCat:    sql.NullInt64{Int64: int64(form.Type), Valid: true},
			ParentTaskID: sql.NullInt64{Int64: int64(taskID), Valid: true},
		},
		)
		if err != nil {
			log.Fatal("AddNote - TaskModel: ", err)
		}
		if noteID != id {
			log.Fatal("AddNote - TaskModel: ", "Note ID and Bridge Note ID do not match")
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

	return func() tea.Msg {
		return SwitchToPreviousViewMsg{}
	}
}

func (m *TaskModel) addTask() tea.Cmd {
	form := &formInput.NewTaskForm{}

	err := form.NewTaskForm()
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		newTask := data.Task{
			Title:    form.TaskTitle,
			Priority: form.Priority,
			Status:   form.Status,
			Archived: form.Archived,
		}

		err = newTask.Create(conn)
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

		m.updateFooter()
		return func() tea.Msg {
			return SwitchToTasksTableViewMsg{}
		}
	}

	return nil
}

func (m *TaskModel) deleteTask() tea.Cmd {
	selectedIDs := []string{}
	ctx := context.Background()
	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[columnKeyID].(string))
	}
	highlightedInfo := m.tableModel.HighlightedRow().Data[columnKeyID].(string)
	highlightedNote := m.tableModel.HighlightedRow().Data[columnKeyNotes].(string)
	taskID, err := strconv.ParseInt(highlightedInfo, 10, 64)
	if err != nil {
		log.Printf("Error converting ID to int64: %s", err)
		return nil
	}

	conn, err := db.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %s", err)
		return nil
	}
	defer conn.Close()

	if len(selectedIDs) <= 1 {
		queries := sqlc.New(conn)
		// query the notes associated with the task
		taskNoteIDs := []int64{}
		taskNotes, err := queries.ReadTaskNote(ctx, sql.NullInt64{Int64: taskID, Valid: true})
		if err != nil {
			log.Printf("Error reading notes: %s", err)
			return nil
		}
		for _, note := range taskNotes {
			taskNoteIDs = append(taskNoteIDs, note.ID)
		}
		// delete those notes
		if highlightedNote != "" {
			_, err := queries.DeleteNotes(ctx, taskNoteIDs)
			if err != nil {
				log.Printf("Error deleting notes: %s", err)
				return nil
			}
		}
		// delete the task
		deletedID, err := queries.DeleteTask(ctx, taskID)
		if err != nil {
			log.Printf("Error deleting task: %s", err)
			return nil
		}
		if deletedID != taskID {
			log.Fatalf("Error deleting task: %s", err)
		} else {
			m.deleteMessage = fmt.Sprintf("You deleted this task:  IDs: %s", highlightedInfo)
		}

	} else if len(selectedIDs) > 1 {
		queries := sqlc.New(conn)
		// query the notes associated with the task
		taskNoteIDs := []int64{}
		taskNotes, err := queries.ReadTaskNote(ctx, sql.NullInt64{Int64: taskID, Valid: true})
		if err != nil {
			log.Printf("Error reading notes: %s", err)
			return nil
		}
		for _, note := range taskNotes {
			taskNoteIDs = append(taskNoteIDs, note.ID)
		}
		// delete those notes
		if highlightedNote != "" {
			_, err := queries.DeleteNotes(ctx, taskNoteIDs)
			if err != nil {
				log.Printf("Error deleting notes: %s", err)
				return nil
			}
		}
		toDeleteTasks := make([]int64, len(selectedIDs))
		for idx, id := range selectedIDs {
			converted_id, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Printf("Error converting ID to int: %s", err)
			}
			toDeleteTasks[idx] = converted_id
		}
		result, err := queries.DeleteTasks(ctx, toDeleteTasks)
		if err != nil {
			log.Printf("Error deleting tasks: %s", err)
			return nil
		}
		if result != int64(len(selectedIDs)) {
			log.Printf("Error deleting tasks - Mismatch between selectedIDs and numDeleted: %s", err)
		}
		m.deleteMessage = fmt.Sprintf("You deleted these tasks:  IDs: %s", strings.Join(selectedIDs, ", "))
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

// View This is where we define the UI for the task table
func (m TaskModel) View() string {
	body := strings.Builder{}

	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Add new task by pressing 'A'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Filter Archived Tasks by pressing 'F'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press left/right or page up/down to move between pages") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press space/enter to select a row, q or ctrl+c to quit") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press D to delete row(s) after selecting them.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press ctrl+n to switch to the Notes View.") + "\n")

	selectedIDs := []string{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[columnKeyID].(string))
	}

	body.WriteString(
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(
				theme.Primary)).
			Render(
				fmt.Sprintf("Selected IDs: %s", strings.Join(selectedIDs, ", "))) + "\n")

	if m.deleteMessage != "" {
		body.WriteString(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(
					theme.Primary)).
				Render(m.deleteMessage) + "\n")
	}

	body.WriteString(m.tableModel.View())
	body.WriteString("\n")

	return body.String()
}

func TaskViewModel() TaskModel {
	theme := tui.GetSelectedTheme()

	columns := []table.Column{
		table.NewColumn(columnKeyID, "ID", 5).WithStyle(
			lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color(theme.Secondary)).
				Align(lipgloss.Center)),
		table.NewFlexColumn(columnKeyTask, "Task", 3),
		table.NewFlexColumn(columnKeyPriority, "Priority", 1),
		table.NewFlexColumn(columnKeyStatus, "Status", 1),
		table.NewFlexColumn(columnKeyArchived, "Archived", 1),
		table.NewFlexColumn(columnKeyTaskAge, "Task Age (Days)", 1),
		table.NewFlexColumn(columnKeyNotes, "Notes", 3),
	}

	model := TaskModel{archiveFilter: true}
	var filteredRows []table.Row
	rows, err := model.loadRowsFromDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range rows {
		archived, ok := row.Data[columnKeyArchived].(sql.NullBool)
		if !ok {
			log.Printf("Error getting archived status from row: %s", err)
		}
		if !archived.Bool {
			filteredRows = append(filteredRows, row)
		}
	}

	keys := table.DefaultKeyMap()
	keys.RowDown.SetKeys("j", "down", "s")
	keys.RowUp.SetKeys("k", "up", "w")
	fmt.Println(filteredRows)

	model.tableModel = table.New(columns).
		WithRows(filteredRows).
		HeaderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)).Bold(true)).
		SelectableRows(true).
		Focused(true).
		Border(customBorder).
		WithKeyMap(keys).
		WithStaticFooter("Footer!").
		WithPageSize(10).
		WithSelectedText(" ", " 󰄲  ").
		WithBaseStyle(
			lipgloss.NewStyle().
				BorderForeground(lipgloss.Color(theme.Primary)).
				Foreground(lipgloss.Color(theme.Success)).
				Align(lipgloss.Left),
		).
		SortByAsc(columnKeyID).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)),
			Data:  "<Missing Data>",
		})

	model.updateFooter()

	return model
}

func (m *TaskModel) updateFooter() {
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

func RunModel(m *TaskModel) {
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func extractNoteTitles(notes []data.Note) string {
	var titles []string
	for _, note := range notes {
		titles = append(titles, note.Title)
	}
	return strings.Join(titles, ", ")
}
