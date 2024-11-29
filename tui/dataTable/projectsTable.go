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
	projectColumnKeyID        = "id"
	projectColumnKeyTask      = "title"
	projectColumnKeyStatus    = "status"
	projectColumnKeyArchived  = "archived"
	projectColumnKeyCreatedAt = "created_at"
	// projectColumnKeyTaskAge   = "age_in_days"
	projectColumnKeyDueDate = "due_date"
	projectColumnKeyNotes   = "notes"
)

// This is the task table "screen" model
type ProjectsModel struct {
	tableModel       table.Model
	totalWidth       int
	totalHeight      int
	horizontalMargin int
	verticalMargin   int
	deleteMessage    string
	archiveFilter    bool
}

// Init initializes the model (can use this to run commands upon model initialization)
func (m *ProjectsModel) Init() tea.Cmd { return nil }

func (m *ProjectsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *ProjectsModel) recalculateTable() {
	m.tableModel = m.tableModel.
		WithTargetWidth(m.calculateWidth()).
		WithMinimumHeight(m.calculateHeight())
}

func (m ProjectsModel) calculateWidth() int {
	return m.totalWidth - m.horizontalMargin
}

func (m ProjectsModel) calculateHeight() int {
	return m.totalHeight - m.verticalMargin - fixedVerticalMargin
}

func (m *ProjectsModel) loadRowsFromDatabase() ([]table.Row, error) {
	ctx := context.Background()
	conn, err := db.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	queries := sqlc.New(conn)
	defer conn.Close()

	projects, err := queries.ReadAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading projects: %w", err)
	}

	var rows []table.Row
	for _, project := range projects {
		row := table.NewRow(table.RowData{
			projectColumnKeyID:       fmt.Sprintf("%d", project.ID),
			projectColumnKeyTask:     project.Title,
			projectColumnKeyStatus:   project.Status.String,
			projectColumnKeyArchived: fmt.Sprintf("%t", project.Archived),
			projectColumnKeyNotes:    project.NoteTitles,
		})
		rows = append(rows, row)
	}
	return rows, nil
}

func (m *ProjectsModel) filterArchives() tea.Cmd {
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
			archived, ok := row.Data[projectColumnKeyArchived]
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
			archived, ok := row.Data[projectColumnKeyArchived]
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

// FIXME: Needs the correct Form here
func (m *ProjectsModel) addNote() tea.Cmd {
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
		projectID, err := strconv.Atoi(highlightedInfo)
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
		id, err := queries.CreateAreaBridgeNote(ctx, sqlc.CreateAreaBridgeNoteParams{
			NoteID:       sql.NullInt64{Int64: int64(noteID), Valid: true},
			ParentCat:    sql.NullInt64{Int64: int64(form.Type), Valid: true},
			ParentAreaID: sql.NullInt64{Int64: int64(projectID), Valid: true},
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

func (m *ProjectsModel) addProject() tea.Cmd {
	// FIXME: Needs the correct form here
	form := &formInput.NewTaskForm{}
	theme := tui.GetSelectedTheme()

	err := form.NewTaskForm(*tui.ThemeGoTask(theme))
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		// TODO: Use SQLC to add, get rid of old version here
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
			return SwitchToProjectsTableViewMsg{}
		}
	}

	return nil
}

func (m *ProjectsModel) deleteProject() tea.Cmd {
	selectedIDs := []string{}
	ctx := context.Background()
	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[projectColumnKeyID].(string))
	}
	highlightedInfo := m.tableModel.HighlightedRow().Data[projectColumnKeyID].(string)
	highlightedNote := m.tableModel.HighlightedRow().Data[projectColumnKeyNotes].(string)
	projectID, err := strconv.ParseInt(highlightedInfo, 10, 64)
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
		projectNoteIDs := []int64{}
		projectNotes, err := queries.ReadAreaNote(ctx, sql.NullInt64{Int64: projectID, Valid: true})
		if err != nil {
			log.Printf("Error reading notes: %s", err)
			return nil
		}
		for _, note := range projectNotes {
			projectNoteIDs = append(projectNoteIDs, note.ID)
		}
		// delete those notes
		if highlightedNote != "" {
			_, err := queries.DeleteNotes(ctx, projectNoteIDs)
			if err != nil {
				log.Printf("Error deleting notes: %s", err)
				return nil
			}
		}
		// delete the project
		deletedID, err := queries.DeleteArea(ctx, projectID)
		if err != nil {
			log.Printf("Error deleting area/project: %s", err)
			return nil
		}
		if deletedID != projectID {
			log.Fatalf("Error deleting project: %s", err)
		} else {
			m.deleteMessage = fmt.Sprintf("You deleted this project:  IDs: %s", highlightedInfo)
		}

	} else if len(selectedIDs) > 1 {
		queries := sqlc.New(conn)
		// query the notes associated with the project
		projectNoteIDs := []int64{}
		taskNotes, err := queries.ReadAreaNote(ctx, sql.NullInt64{Int64: projectID, Valid: true})
		if err != nil {
			log.Printf("Error reading notes: %s", err)
			return nil
		}
		for _, note := range taskNotes {
			projectNoteIDs = append(projectNoteIDs, note.ID)
		}
		// delete those notes
		if highlightedNote != "" {
			_, err := queries.DeleteNotes(ctx, projectNoteIDs)
			if err != nil {
				log.Printf("Error deleting notes: %s", err)
				return nil
			}
		}
		toDeleteProjects := make([]int64, len(selectedIDs))
		for idx, id := range selectedIDs {
			converted_id, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Printf("Error converting ID to int: %s", err)
			}
			toDeleteProjects[idx] = converted_id
		}
		result, err := queries.DeleteTasks(ctx, toDeleteProjects)
		if err != nil {
			log.Printf("Error deleting projects: %s", err)
			return nil
		}
		if result != int64(len(selectedIDs)) {
			log.Printf("Error deleting projectss - Mismatch between selectedIDs and numDeleted: %s", err)
		}
		m.deleteMessage = fmt.Sprintf("You deleted these projects:  IDs: %s", strings.Join(selectedIDs, ", "))
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

// View This is where we define the UI for the projects table
func (m ProjectsModel) View() string {
	body := strings.Builder{}

	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Add new Project by pressing 'A'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Filter Archived Projects by pressing 'F'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press left/right or page up/down to move between pages") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press space/enter to select a row, q or ctrl+c to quit") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press D to delete row(s) after selecting them.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press ctrl+n to switch to the Notes View.") + "\n")

	selectedIDs := []string{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[projectColumnKeyID].(string))
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

func ProjectViewModel() ProjectsModel {
	theme := tui.GetSelectedTheme()

	columns := []table.Column{
		table.NewColumn(projectColumnKeyID, "ID", 5).WithStyle(
			lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color(theme.Secondary)).
				Align(lipgloss.Center)),
		table.NewFlexColumn(projectColumnKeyTask, "Project", 3),
		table.NewFlexColumn(projectColumnKeyStatus, "Status", 1),
		table.NewFlexColumn(projectColumnKeyArchived, "Archived", 1),
		// table.NewFlexColumn(columnKeyTaskAge, "Project Age (Days)", 1),
		table.NewFlexColumn(projectColumnKeyNotes, "Notes", 3),
	}

	model := ProjectsModel{archiveFilter: true}
	var filteredRows []table.Row
	rows, err := model.loadRowsFromDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range rows {
		archived, ok := row.Data[projectColumnKeyArchived]
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
		SortByAsc(projectColumnKeyID).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)),
			Data:  "<Missing Data>",
		})

	model.updateFooter()

	return model
}

func (m *ProjectsModel) updateFooter() {
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

func RunProjectsModel(m *ProjectsModel) {
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
