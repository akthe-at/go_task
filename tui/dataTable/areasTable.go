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

	"github.com/evertras/bubble-table/table"
)

const (
	areaColumnKeyID        = "id"
	areaColumnKeyProject   = "title"
	areaColumnKeyStatus    = "status"
	areaColumnKeyArchived  = "archived"
	areaColumnKeyCreatedAt = "created_at"
	areaColumnKeyNotes     = "notes"
	areaColumnKeyPath      = "path"
)

// This is the task table "screen" model
type AreasModel struct {
	tableModel           table.Model
	totalWidth           int
	totalHeight          int
	horizontalMargin     int
	verticalMargin       int
	deleteMessage        string
	archiveFilterEnabled bool
	rowFilter            bool
}

// Init initializes the model (can use this to run commands upon model initialization)
func (m *AreasModel) Init() tea.Cmd { return nil }

func (m *AreasModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *AreasModel) recalculateTable() {
	m.tableModel = m.tableModel.
		WithTargetWidth(m.calculateWidth()).
		WithMinimumHeight(m.calculateHeight())
}

func (m AreasModel) calculateWidth() int {
	return m.totalWidth - m.horizontalMargin
}

func (m AreasModel) calculateHeight() int {
	return m.totalHeight - m.verticalMargin - fixedVerticalMargin
}

func (m *AreasModel) loadRowsFromDatabase() ([]table.Row, error) {
	ctx := context.Background()
	conn, err := db.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	queries := sqlc.New(conn)
	defer conn.Close()

	areas, err := queries.ReadAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading areas: %w", err)
	}

	var rows []table.Row
	for _, area := range areas {
		formattedPath := path.Base(area.Path.String)
		if formattedPath == "." {
			formattedPath = ""
		}
		row := table.NewRow(table.RowData{
			areaColumnKeyID:       fmt.Sprintf("%d", area.ID),
			areaColumnKeyProject:  area.Title,
			areaColumnKeyStatus:   area.Status.String,
			areaColumnKeyArchived: fmt.Sprintf("%t", area.Archived),
			areaColumnKeyNotes:    area.NoteTitles,
			areaColumnKeyPath:     formattedPath,
		})
		rows = append(rows, row)
	}
	filteredRows := []table.Row{}
	for _, row := range rows {
		archived, ok := row.Data[areaColumnKeyArchived]
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

func (m *AreasModel) filterRows() tea.Cmd {
	ctx := context.Background()
	dbConn, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("There was an issue connecting to the database: %s", err)
	}
	defer dbConn.Close()
	queries := sqlc.New(dbConn)
	m.rowFilter = !m.rowFilter

	if m.tableModel.HighlightedRow().Data[areaColumnKeyID] != nil {
		rowID, err := strconv.ParseInt(m.tableModel.HighlightedRow().Data[areaColumnKeyID].(string), 10, 64)
		if err != nil {
			log.Fatalf("Error converting ID to int: %s", err)
		}

		result, err := queries.ReadArea(ctx, rowID)
		if err != nil {
			slog.Error("Error reading area from database: %s", "error", err)
			return nil
		}

		var rows []table.Row
		row := table.NewRow(table.RowData{
			areaColumnKeyID:       fmt.Sprintf("%d", result.ID),
			areaColumnKeyProject:  result.Title,
			areaColumnKeyStatus:   result.Status.String,
			areaColumnKeyArchived: fmt.Sprintf("%t", result.Archived),
			areaColumnKeyNotes:    fmt.Sprintf("%v", result.Title_2),
		})
		rows = append(rows, row)

		m.tableModel = m.tableModel.WithRows(rows)
		m.updateFooter()

		return nil
	}
	return nil
}

func (m *AreasModel) filterArchives() tea.Cmd {
	m.archiveFilterEnabled = !m.archiveFilterEnabled
	m.refreshTableData()

	m.updateFooter()
	return nil
}

func (m *AreasModel) updateStatus(newStatus data.StatusType) tea.Cmd {
	var selectedIDs []int64
	ctx := context.Background()
	for _, row := range m.tableModel.SelectedRows() {
		convertedID, err := strconv.ParseInt(row.Data[columnKeyID].(string), 10, 64)
		if err != nil {
			log.Fatalf("AreasModel - UpdateStatus: Error converting ID to int64: %v", err)
		}
		selectedIDs = append(selectedIDs, convertedID)
	}

	conn, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("AreasModel - UpdateStatus: Error connecting to database: %v", err)
	}
	defer conn.Close()

	queries := sqlc.New(conn)

	if len(selectedIDs) < 1 {
		highlightedInfo := m.tableModel.HighlightedRow().Data[columnKeyID].(string)
		taskID, err := strconv.ParseInt(highlightedInfo, 10, 64)
		if err != nil {
			log.Fatalf("AreasModel - UpdateStatus: Error converting ID to int64: %v", err)
		}

		_, err = queries.UpdateAreaStatus(ctx, sqlc.UpdateAreaStatusParams{Status: sql.NullString{String: string(newStatus), Valid: true}, ID: taskID})
		if err != nil {
			log.Fatalf("AreasModel - UpdateStatus: Error updating Area status: %v", err)
		}
	} else if len(selectedIDs) >= 1 {
		for _, ID := range selectedIDs {
			_, err := queries.UpdateAreaStatus(ctx, sqlc.UpdateAreaStatusParams{Status: sql.NullString{String: string(newStatus), Valid: true}, ID: ID})
			if err != nil {
				log.Fatalf("AreasModel - UpdateStatus: Error updating area status: %v", err)
			}
		}
	}

	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Fatalf("Error loading rows from database: %s", err)
	}

	m.tableModel = m.tableModel.WithRows(rows)
	m.updateFooter()

	return nil
}

func (m *AreasModel) addNote() tea.Cmd {
	form := &formInput.NewNoteForm{}
	theme := tui.GetSelectedTheme()
	err := form.NewNoteForm(*tui.ThemeGoTask(theme))
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		selectedIDs := []string{}
		var areaID int

		for _, row := range m.tableModel.SelectedRows() {
			selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(string))
		}

		if len(selectedIDs) == 1 {
			areaID, err = strconv.Atoi(selectedIDs[0])
			if err != nil {
				log.Printf("Error converting ID to int: %s", err)
				return nil
			}
		} else {
			highlightedInfo := fmt.Sprintf("%v", m.tableModel.HighlightedRow().Data[NoteColumnKeyID])
			areaID, err = strconv.Atoi(highlightedInfo)
			if err != nil {
				log.Printf("Error converting ID to int: %s", err)
				return nil
			}
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
		id, err := queries.CreateAreaBridgeNote(ctx, sqlc.CreateAreaBridgeNoteParams{
			NoteID:       sql.NullInt64{Int64: int64(noteID), Valid: true},
			ParentCat:    sql.NullInt64{Int64: int64(form.Type), Valid: true},
			ParentAreaID: sql.NullInt64{Int64: int64(areaID), Valid: true},
		},
		)
		if err != nil {
			log.Fatal("AddNote - ProjectsModel: ", err)
		}
		if noteID != id {
			log.Fatal("AddNote - ProjectsModel: ", "Note ID and Bridge Note ID do not match")
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

func (m *AreasModel) addArea() tea.Cmd {
	form := &formInput.NewAreaForm{}
	theme := tui.GetSelectedTheme()

	err := form.NewAreaForm(*tui.ThemeGoTask(theme))
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)
		newArea := sqlc.CreateAreaParams{
			Title:    form.AreaTitle,
			Status:   sql.NullString{String: string(form.Status), Valid: true},
			Archived: form.Archived,
		}

		result, err := queries.CreateArea(ctx, newArea)
		if err != nil {
			log.Fatalf("Error creating new area: %v", err)
		}

		var projectID int64
		if form.ProjectAssignment == "local" {
			projID, err := queries.CheckProgProjectExists(ctx, form.ProgProject)
			if err != nil {
				log.Fatalf("Error checking if project exists: %v", err)
			}
			switch projID {
			case 0:
				projectID, err = queries.InsertProgProject(ctx, form.ProgProject)
				if err != nil {
					log.Fatalf("Error inserting project: %v", err)
				}
			case 1:
				projectID = projID
			default:
				log.Fatalf("Unexpected projID: %v", projID)
			}
			err = queries.CreateProjectTaskLink(ctx,
				sqlc.CreateProjectTaskLinkParams{
					ProjectID:    sql.NullInt64{Int64: projectID, Valid: true},
					ParentCat:    sql.NullInt64{Int64: int64(data.AreaNoteType), Valid: true},
					ParentTaskID: sql.NullInt64{Int64: result, Valid: true},
				},
			)
			if err != nil {
				log.Fatalf("Error inserting project link: %v", err)
			}
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
			return SwitchToProjectsTableViewMsg{}
		}
	}

	return nil
}

func (m *AreasModel) refreshTableData() {
	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Printf("Error loading rows from database: %s", err)
	}

	m.tableModel = m.tableModel.WithRows(rows)

	m.updateFooter()
}

func (m *AreasModel) addTaskToArea() tea.Cmd {
	form := &formInput.NewTaskForm{}
	theme := tui.GetSelectedTheme()

	selectedAreaIDs := []string{}
	for _, row := range m.tableModel.SelectedRows() {
		selectedAreaIDs = append(selectedAreaIDs, row.Data[areaColumnKeyID].(string))
	}
	if len(selectedAreaIDs) > 1 {
		log.Fatal("You can only select one area at a time to add a task to.")
	}
	highlightedInfo := m.tableModel.HighlightedRow().Data[areaColumnKeyID].(string)
	areaID, err := strconv.ParseInt(highlightedInfo, 10, 64)
	if err != nil {
		log.Fatalf("Error converting ID to int64: %s", err)
	}

	err = form.NewTaskForm(*tui.ThemeGoTask(theme))
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		queries := sqlc.New(conn)
		_, err = queries.CreateTask(ctx, sqlc.CreateTaskParams{
			Title:    form.TaskTitle,
			Priority: sql.NullString{String: string(form.Priority), Valid: true},
			Status:   sql.NullString{String: string(form.Status), Valid: true},
			Archived: form.Archived,
			AreaID:   sql.NullInt64{Int64: areaID, Valid: true},
		})
		if err != nil {
			log.Fatalf("Error creating new task: %v", err)
		}

		// Requery the database and update the table model
		rows, err := m.loadRowsFromDatabase()
		if err != nil {
			log.Printf("Error loading rows from database: %s", err)
			return nil
		}
		m.tableModel = m.tableModel.WithRows(rows)

		m.updateFooter()
		m.recalculateTable()
		// FIXME: theres a visual bug here after completing this!
		//
		return func() tea.Msg {
			return SwitchToProjectsTableViewMsg{}
		}
	}

	return nil
}

func (m *AreasModel) deleteArea() tea.Cmd {
	selectedIDs := []string{}
	ctx := context.Background()
	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[areaColumnKeyID].(string))
	}
	highlightedInfo := m.tableModel.HighlightedRow().Data[areaColumnKeyID].(string)
	highlightedNote := m.tableModel.HighlightedRow().Data[areaColumnKeyNotes].(string)
	areaID, err := strconv.ParseInt(highlightedInfo, 10, 64)
	if err != nil {
		log.Printf("Error converting ID to int64: %s", err)
		return nil
	}

	conn, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	defer conn.Close()

	if len(selectedIDs) <= 1 {
		queries := sqlc.New(conn)
		// query the notes associated with the task
		areaNoteIDs := []int64{}
		areaNotes, err := queries.ReadAreaNote(ctx, sql.NullInt64{Int64: areaID, Valid: true})
		if err != nil {
			log.Fatalf("Error reading notes: %s", err)
		}
		for _, note := range areaNotes {
			areaNoteIDs = append(areaNoteIDs, note.ID)
		}
		// delete those notes
		if highlightedNote != "" {
			_, err := queries.DeleteNotes(ctx, areaNoteIDs)
			if err != nil {
				log.Fatalf("Error deleting notes: %s", err)
			}
		}
		// delete the project
		deletedID, err := queries.DeleteSingleArea(ctx, areaID)
		if err != nil {
			log.Fatalf("Error deleting area: %s", err)
		}
		if deletedID != areaID {
			log.Fatalf("Error deleting area: %s", err)
		} else {
			m.deleteMessage = fmt.Sprintf("You deleted the following area:  IDs: %s", highlightedInfo)
		}

	} else if len(selectedIDs) > 1 {
		queries := sqlc.New(conn)
		// query the notes associated with the area
		areaNoteIDs := []int64{}
		taskNotes, err := queries.ReadAreaNote(ctx, sql.NullInt64{Int64: areaID, Valid: true})
		if err != nil {
			log.Fatalf("Error reading notes: %s", err)
		}
		for _, note := range taskNotes {
			areaNoteIDs = append(areaNoteIDs, note.ID)
		}
		// delete those notes
		if highlightedNote != "" {
			_, err := queries.DeleteNotes(ctx, areaNoteIDs)
			if err != nil {
				log.Fatalf("Error deleting notes: %s", err)
			}
		}
		areasToDelete := make([]int64, len(selectedIDs))
		for idx, id := range selectedIDs {
			converted_id, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Printf("Error converting ID to int: %s", err)
			}
			areasToDelete[idx] = converted_id
		}
		result, err := queries.DeleteTasks(ctx, areasToDelete)
		if err != nil {
			log.Fatalf("Error deleting areas: %s", err)
		}
		if result != int64(len(selectedIDs)) {
			log.Fatalf("Error deleting areas - Mismatch between selectedIDs and numDeleted: %s", err)
		}
		m.deleteMessage = fmt.Sprintf("You deleted these areas:  IDs: %s", strings.Join(selectedIDs, ", "))
	}

	// Requery the database and update the table model
	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Fatalf("Error loading rows from database: %s", err)
	}
	m.tableModel = m.tableModel.WithRows(rows)

	m.updateFooter()

	return nil
}

func (m AreasModel) View() string {
	body := strings.Builder{}

	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Add a new Area by pressing 'A'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Filter Archived Areas by pressing 'F'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press left/right or page up/down to move between pages") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press space/enter to select a row, q or ctrl+c to quit") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press D to delete row(s) after selecting them.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press ctrl+n to switch to the Notes View.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press ctrl+t to switch to the Tasks View.") + "\n")
	selectedIDs := []string{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[areaColumnKeyID].(string))
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

func AreaViewModel() AreasModel {
	theme := tui.GetSelectedTheme()

	columns := []table.Column{
		table.NewColumn(areaColumnKeyID, "ID", 10).WithStyle(
			lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color(theme.Secondary)).
				Align(lipgloss.Center)),
		table.NewFlexColumn(areaColumnKeyProject, "Area", 3),
		table.NewFlexColumn(areaColumnKeyStatus, "Status", 1),
		table.NewFlexColumn(areaColumnKeyArchived, "Archived", 1),
		table.NewFlexColumn(areaColumnKeyPath, "Repo", 1),
		table.NewFlexColumn(areaColumnKeyNotes, "Notes", 3),
	}

	model := AreasModel{archiveFilterEnabled: true}
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
		SortByAsc(areaColumnKeyID).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)),
			Data:  "<Missing Data>",
		})

	model.updateFooter()

	return model
}

func (m *AreasModel) updateFooter() {
	highlightedRow := m.tableModel.HighlightedRow()
	rowID, ok := highlightedRow.Data[areaColumnKeyID]
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

func RunProjectsModel(m *AreasModel) {
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m *AreasModel) archiveArea() tea.Cmd {
	selectedIDs := make(map[int64]bool)
	var currentArchiveState bool
	ctx := context.Background()
	for _, row := range m.tableModel.SelectedRows() {
		convertedID, err := strconv.ParseInt(row.Data[areaColumnKeyID].(string), 10, 64)
		if err != nil {
			slog.Error("AreasModel - archiveArea: Error converting ID to int64: %v", "error", err)
		}
		parsedArchiveStatus, err := strconv.ParseBool(row.Data[areaColumnKeyArchived].(string))
		if err != nil {
			slog.Error("AreasModel - archiveArea: Error converting archived status to bool: %v", "error", err)
		}
		selectedIDs[convertedID] = parsedArchiveStatus
	}

	conn, err := db.ConnectDB()
	if err != nil {
		slog.Error("AreasModel - archiveArea: Error connecting to database: %v", "error", err)
		return nil
	}
	defer conn.Close()

	queries := sqlc.New(conn)

	if len(selectedIDs) < 1 {
		highlightedInfo := m.tableModel.HighlightedRow().Data[areaColumnKeyID].(string)
		currentArchiveState, err = strconv.ParseBool(m.tableModel.HighlightedRow().Data[areaColumnKeyArchived].(string))
		if err != nil {
			slog.Error("AreasModel - archiveArea: Error converting archived status to bool: %v", "error", err)
		}
		taskID, err := strconv.ParseInt(highlightedInfo, 10, 64)
		if err != nil {
			slog.Error("AreasModel - archiveArea: Error converting ID to int64: %v", "error", err)
			return nil
		}
		_, err = queries.UpdateAreaArchived(ctx, sqlc.UpdateAreaArchivedParams{
			Archived: !currentArchiveState,
			ID:       taskID,
		})
		if err != nil {
			slog.Error("AreasModel - archiveArea: Error updating area archived status: %v", "error", err)
		}

	} else if len(selectedIDs) >= 1 {
		for ID, archiveStatus := range selectedIDs {
			_, err = queries.UpdateAreaArchived(ctx, sqlc.UpdateAreaArchivedParams{
				Archived: !archiveStatus,
				ID:       ID,
			})
			if err != nil {
				slog.Error("AreasModel - archiveArea: Error updating area archived status: %v", "error", err)
			}
		}
	}

	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		slog.Error("AreasModel - archiveArea: Error loading rows from database: %v", "error", err)
		return nil
	}

	m.tableModel = m.tableModel.WithRows(rows)
	m.updateFooter()

	return nil
}
