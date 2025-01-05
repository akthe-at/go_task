package datatable

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
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
	columnKeyPath       = "path"
	columnKeyArea       = "parent_area"
	minWidth            = 120
	minHeight           = 10
	fixedVerticalMargin = 80
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
	deleteMessage        string
	tableModel           table.Model
	totalWidth           int
	totalHeight          int
	horizontalMargin     int
	verticalMargin       int
	selectedRowID        int
	archiveFilterEnabled bool
	rowFilter            bool
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
		case "enter":
			cmds = append(cmds, m.filterRows())
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

func (m *TaskModel) refreshTableData() {
	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Printf("Error loading rows from database: %s", err)
	}

	m.tableModel = m.tableModel.WithRows(rows)
	m.updateFooter()
}

func (m TaskModel) calculateWidth() int {
	return m.totalWidth - m.horizontalMargin
}

func (m TaskModel) calculateHeight() int {
	return m.totalHeight - m.verticalMargin - fixedVerticalMargin
}

func (m *TaskModel) loadRowsFromDatabase() ([]table.Row, error) {
	ctx := context.Background()
	conn, _, err := db.ConnectDB()
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
		formattedPath := path.Base(task.Path.String)
		if formattedPath == "." {
			formattedPath = ""
		}
		row := table.NewRow(table.RowData{
			columnKeyID:       fmt.Sprintf("%d", task.ID),
			columnKeyTask:     task.Title,
			columnKeyPriority: task.Priority.String,
			columnKeyStatus:   task.Status.String,
			columnKeyArchived: fmt.Sprintf("%t", task.Archived),
			columnKeyTaskAge:  fmt.Sprintf("%v Days", task.AgeInDays),
			columnKeyNotes:    task.NoteTitles,
			columnKeyPath:     formattedPath,
			columnKeyArea:     task.ParentArea.String,
		})
		rows = append(rows, row)

	}

	filteredRows := []table.Row{}
	for _, row := range rows {
		archived, ok := row.Data[columnKeyArchived]
		if !ok {
			log.Printf("Error getting archived status from row: %s", err)
			return nil, err
		}
		if m.archiveFilterEnabled && archived == "false" {
			filteredRows = append(filteredRows, row)
		} else if !m.archiveFilterEnabled {
			filteredRows = append(filteredRows, row)
		}
	}

	return filteredRows, nil
}

func (m *TaskModel) filterRows() tea.Cmd {
	ctx := context.Background()
	dbConn, _, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("There was an issue connecting to the database: %s", err)
	}
	defer dbConn.Close()
	queries := sqlc.New(dbConn)
	m.rowFilter = !m.rowFilter

	if m.tableModel.HighlightedRow().Data[columnKeyID] != nil {
		rowID, err := strconv.ParseInt(m.tableModel.HighlightedRow().Data[columnKeyID].(string), 10, 64)
		if err != nil {
			log.Fatalf("Error converting ID to int: %s", err)
		}

		result, err := queries.ReadTask(ctx, rowID)
		if err != nil {
			log.Printf("Error reading task from database: %s", err)
			return nil
		}

		var rows []table.Row
		row := table.NewRow(table.RowData{
			columnKeyID:       fmt.Sprintf("%d", result.TaskID),
			columnKeyTask:     result.TaskTitle,
			columnKeyPriority: result.Priority.String,
			columnKeyStatus:   result.Status.String,
			columnKeyArchived: fmt.Sprintf("%t", result.Archived),
			columnKeyTaskAge:  fmt.Sprintf("%v Days", result.AgeInDays),
			columnKeyNotes:    fmt.Sprintf("%v", result.NoteTitle),
			columnKeyPath:     result.ProgProj.String,
			columnKeyArea:     result.ParentArea.String,
		})
		rows = append(rows, row)

		m.tableModel = m.tableModel.WithRows(rows)
		m.updateFooter()

		return nil
	}
	return nil
}

func (m *TaskModel) filterArchives() tea.Cmd {
	m.archiveFilterEnabled = !m.archiveFilterEnabled
	m.refreshTableData()

	m.updateFooter()
	return nil
}

func (m *TaskModel) addNote() tea.Cmd {
	form := &formInput.NewNoteForm{}
	theme := tui.GetSelectedTheme()
	err := form.NewNoteForm(*tui.ThemeGoTask(theme))
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		selectedIDs := []string{}

		for _, row := range m.tableModel.SelectedRows() {
			selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(string))
		}
		if len(selectedIDs) > 1 {
			log.Fatal("Currently unable to add note to multiple tasks at once")
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
		conn, _, err := db.ConnectDB()
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
			NoteID:       sql.NullInt64{Int64: int64(noteID), Valid: true},
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
	theme := tui.GetSelectedTheme()

	err := form.NewTaskForm(*tui.ThemeGoTask(theme))
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		ctx := context.Background()
		conn, _, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		queries := sqlc.New(conn)

		newTaskID, err := queries.GetTaskID(ctx)
		if err != nil && err != sql.ErrNoRows {
			log.Fatalf("addTask - TaskModel: Error getting task ID: %v", err)
		}
		if err == sql.ErrNoRows {
			newTaskID, err = queries.NoTaskIDs(ctx)
			if err != nil {
				log.Fatalf("Failed to find the next available task ID: %v", err)
			} else {
				_, err = queries.DeleteTaskID(ctx, newTaskID)
				if err != nil {
					log.Fatalf("Error deleting task ID: %v", err)
				}
			}
		}

		newTask := sqlc.CreateTaskParams{
			ID:       newTaskID,
			Title:    form.TaskTitle,
			Priority: sql.NullString{String: string(form.Priority), Valid: true},
			Status:   sql.NullString{String: string(form.Status), Valid: true},
			Archived: form.Archived,
		}

		result, err := queries.CreateTask(ctx, newTask)
		if err != nil {
			log.Fatalf("Error creating task: %v", err)
		}

		if form.AreaAssignment == "yes" {
			areaID, err := strconv.ParseInt(form.Area, 10, 64)
			if err != nil {
				log.Fatalf("Error parsing area ID: %v", err)
			}
			fmt.Println("AreaID Post Conversion is: ", areaID)
			_, err = queries.UpdateTaskArea(ctx, sqlc.UpdateTaskAreaParams{
				AreaID: sql.NullInt64{Int64: areaID, Valid: true}, ID: result,
			})
			if err != nil {
				slog.Error("Error updating task area: %v", "error", err)
			}
		}

		var projectID int64
		if form.ProjectAssignment == "local" {
			projID, err := queries.CheckProgProjectExists(ctx, form.ProgProject)
			if err != nil {
				log.Fatalf("Error checking if project exists: %v", err)
			}
			switch {
			case projID == 0:
				projectID, err = queries.InsertProgProject(ctx, form.ProgProject)
				if err != nil {
					log.Fatalf("Error inserting project: %v", err)
				}
			case projID > 0:
				projectID = projID
			default:
				log.Fatalf("Unexpected projID: %v", projID)
			}
			err = queries.CreateProjectTaskLink(ctx,
				sqlc.CreateProjectTaskLinkParams{
					ProjectID:    sql.NullInt64{Int64: projectID, Valid: true},
					ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
					ParentTaskID: sql.NullInt64{Int64: result, Valid: true},
				},
			)
			if err != nil {
				log.Fatalf("Error inserting project link: %v", err)
			}
		}

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

func (m *TaskModel) togglePriorityStatus() tea.Cmd {
	selectedIDs := make(map[int64]data.PriorityType)
	ctx := context.Background()
	conn, _, err := db.ConnectDB()
	if err != nil {
		slog.Error("TaskModel - togglePriorityStatus: Error connecting to database: %v", "error", err)
		return nil
	}
	defer conn.Close()

	queries := sqlc.New(conn)

	if len(selectedIDs) < 1 {
		highlightedInfo := m.tableModel.HighlightedRow().Data[columnKeyID].(string)
		currentPriorityStateStr, ok := m.tableModel.HighlightedRow().Data[columnKeyPriority].(string)
		if !ok {
			slog.Error("TaskModel - togglePriorityStatus: Error converting priority status to PriorityType: %v", "error", err)
			return nil
		}

		currentPriorityState, err := data.StringToPriorityType(currentPriorityStateStr)
		if err != nil {
			slog.Error("TaskModel - togglePriorityStatus: Error converting priority status to PriorityType: %v", "error", err)
			return nil
		}

		var newPriorityState data.PriorityType
		switch currentPriorityState {
		case data.PriorityTypeLow:
			newPriorityState = data.PriorityTypeMedium
		case data.PriorityTypeMedium:
			newPriorityState = data.PriorityTypeHigh
		case data.PriorityTypeHigh:
			newPriorityState = data.PriorityTypeUrgent
		case data.PriorityTypeUrgent:
			newPriorityState = data.PriorityTypeLow
		}

		taskID, err := strconv.ParseInt(highlightedInfo, 10, 64)
		if err != nil {
			slog.Error("TaskModel - togglePriorityStatus: Error converting ID to int64: %v", "error", err)
			return nil
		}
		queries.UpdateTaskPriority(ctx, sqlc.UpdateTaskPriorityParams{
			Priority: sql.NullString{String: string(newPriorityState), Valid: true},
			ID:       taskID,
		},
		)
	}

	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		slog.Error("TaskModel - togglePriorityStatus: Error loading rows from database: %v", "error", err)
		return nil
	}

	m.tableModel = m.tableModel.WithRows(rows)
	m.updateFooter()

	return nil
}

func (m *TaskModel) archiveTask() tea.Cmd {
	selectedIDs := make(map[int64]bool)
	var currentArchiveState bool
	ctx := context.Background()
	for _, row := range m.tableModel.SelectedRows() {
		convertedID, err := strconv.ParseInt(row.Data[columnKeyID].(string), 10, 64)
		if err != nil {
			slog.Error("TaskModel - archiveTask: Error converting ID to int64: %v", "error", err)
		}
		parsedArchiveStatus, err := strconv.ParseBool(row.Data[columnKeyArchived].(string))
		if err != nil {
			slog.Error("TaskModel - archiveTask: Error converting archived status to bool: %v", "error", err)
		}
		selectedIDs[convertedID] = parsedArchiveStatus
	}

	conn, _, err := db.ConnectDB()
	if err != nil {
		slog.Error("TaskModel - archiveTask: Error connecting to database: %v", "error", err)
		return nil
	}
	defer conn.Close()

	queries := sqlc.New(conn)

	if len(selectedIDs) < 1 {
		highlightedInfo := m.tableModel.HighlightedRow().Data[columnKeyID].(string)
		currentArchiveState, err = strconv.ParseBool(m.tableModel.HighlightedRow().Data[columnKeyArchived].(string))
		if err != nil {
			slog.Error("TaskModel - archiveTask: Error converting archived status to bool: %v", "error", err)
		}
		taskID, err := strconv.ParseInt(highlightedInfo, 10, 64)
		if err != nil {
			slog.Error("TaskModel - archiveTask: Error converting ID to int64: %v", "error", err)
			return nil
		}
		_, err = queries.UpdateTaskArchived(ctx, sqlc.UpdateTaskArchivedParams{
			Archived: !currentArchiveState,
			ID:       taskID,
		})
		if err != nil {
			slog.Error("TaskModel - archiveTask: Error updating task archived status: %v", "error", err)
		}

	} else if len(selectedIDs) >= 1 {
		for ID, archiveStatus := range selectedIDs {
			_, err = queries.UpdateTaskArchived(ctx, sqlc.UpdateTaskArchivedParams{
				Archived: !archiveStatus,
				ID:       ID,
			})
			if err != nil {
				slog.Error("TaskModel - archiveTask: Error updating task archived status: %v", "error", err)
			}
		}
	}

	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		slog.Error("TaskModel - archiveTask: Error loading rows from database: %v", "error", err)
		return nil
	}

	m.tableModel = m.tableModel.WithRows(rows)
	m.updateFooter()

	return nil
}

func (m *TaskModel) updateStatus(newStatus data.StatusType) tea.Cmd {
	var selectedIDs []int64
	ctx := context.Background()
	for _, row := range m.tableModel.SelectedRows() {
		convertedID, err := strconv.ParseInt(row.Data[columnKeyID].(string), 10, 64)
		if err != nil {
			slog.Error("TaskModel - UpdateStatus: Error converting ID to int64: %v", "error", err)
		}
		selectedIDs = append(selectedIDs, convertedID)
	}

	conn, _, err := db.ConnectDB()
	if err != nil {
		slog.Error("TaskModel - UpdateStatus: Error connecting to database: %v", "error", err)
		return nil
	}
	defer conn.Close()

	queries := sqlc.New(conn)

	if len(selectedIDs) < 1 {
		highlightedInfo := m.tableModel.HighlightedRow().Data[columnKeyID].(string)
		taskID, err := strconv.ParseInt(highlightedInfo, 10, 64)
		if err != nil {
			slog.Error("TaskModel - UpdateStatus: Error converting ID to int64: %v", "error", err)
			return nil
		}

		_, err = queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{Status: sql.NullString{String: string(newStatus), Valid: true}, ID: taskID})
		if err != nil {
			slog.Error("TaskModel - UpdateStatus: Error updating task status: %v", "error", err)
			return nil
		}
	} else if len(selectedIDs) >= 1 {
		for _, ID := range selectedIDs {
			_, err := queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{Status: sql.NullString{String: string(newStatus), Valid: true}, ID: ID})
			if err != nil {
				slog.Error("TaskModel - UpdateStatus: Error updating task status: %v", "error", err)
				return nil
			}
		}
	}

	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Printf("Error loading rows from database: %s", err)
		return nil
	}

	m.tableModel = m.tableModel.WithRows(rows)
	m.updateFooter()

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

	conn, _, err := db.ConnectDB()
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
		_, err = queries.RecycleTaskID(ctx, deletedID)
		if err != nil {
			log.Fatalf("Error recycling task ID: %s", err)
		}
		if deletedID != taskID {
			log.Fatalf("Error deleting task: %s", err)
		} else {
			m.deleteMessage = fmt.Sprintf("You deleted the following task: %s", highlightedInfo)
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
		for _, taskID := range toDeleteTasks {
			_, err := queries.DeleteTask(ctx, taskID)
			if err != nil {
				log.Fatalf("Error deleting task: %v", err)
			}
			_, err = queries.RecycleTaskID(ctx, taskID)
			if err != nil {
				log.Fatalf("Error recycling task ID: %v", err)
			}
		}
		m.deleteMessage = fmt.Sprintf("You deleted the following tasks: %s", strings.Join(selectedIDs, ", "))
	}

	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Printf("Error loading rows from database: %s", err)
		return nil
	}
	m.tableModel = m.tableModel.WithRows(rows)
	m.updateFooter()

	return nil
}

// View This is where we define the UI for the task table
func (m TaskModel) View() string {
	body := strings.Builder{}

	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Add a new Task by pressing 'A'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Filter Archived Tasks by pressing 'F'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press left/right or page up/down to move between pages") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press 'space' to select a row, 'q' or 'ctrl+c' to quit") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press 'enter' to filter table to a singular task") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press 'backspace' to delete row(s) after selecting or highlighting them.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press 't'/'p'/'d'/'D' to Toggle Task Status to Todo, Planning, Doing, or Done, respectively.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press 'P' to toggle the priority status of a highlighted task.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press 'a' to toggle archive status of a highlighted or selected tasks.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press 'ctrl+n' to switch to the Notes View.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press 'ctrl+p' to switch to the Areas View.") + "\n")

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
		table.NewColumn(columnKeyPriority, "Priority", 10),
		table.NewColumn(columnKeyStatus, "Status", 10),
		table.NewColumn(columnKeyArchived, "Archived", 10),
		table.NewColumn(columnKeyTaskAge, "Task Age", 15),
		table.NewFlexColumn(columnKeyNotes, "Notes", 3),
		table.NewFlexColumn(columnKeyPath, "Repo", 1),
		table.NewFlexColumn(columnKeyArea, "Area", 3),
	}

	model := TaskModel{archiveFilterEnabled: true, rowFilter: false}

	rows, err := model.loadRowsFromDatabase()
	if err != nil {
		log.Fatal(err)
	}

	keys := table.DefaultKeyMap()
	keys.RowDown.SetKeys("j", "down", "s")
	keys.RowUp.SetKeys("k", "up", "w")

	model.tableModel = table.New(columns).
		WithRows(rows).
		HeaderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)).Bold(true)).
		SelectableRows(true).
		Focused(true).
		Border(customBorder).
		WithKeyMap(keys).
		WithStaticFooter("Footer!").
		WithPageSize(50).
		WithSelectedText(" ", " 󰄲  ").
		WithBaseStyle(
			lipgloss.NewStyle().
				BorderForeground(lipgloss.Color(theme.Primary)).
				Foreground(lipgloss.Color(theme.Success)).
				Align(lipgloss.Left),
		).
		SortByDesc(columnKeyTaskAge).
		WithMultiline(true).
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

func extractNoteTitles(notes []sqlc.Note) string {
	var titles []string
	for _, note := range notes {
		titles = append(titles, note.Title)
	}
	return strings.Join(titles, ", ")
}
