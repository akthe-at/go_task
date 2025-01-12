package datatable

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/akthe-at/go_task/config"
	data "github.com/akthe-at/go_task/data"
	db "github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui"
	"github.com/akthe-at/go_task/tui/formInput"
	"github.com/akthe-at/go_task/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	NoteColumnKeyID      = "id"
	NoteColumnKey        = "title"
	NoteColumnPath       = "path"
	NoteColumnLink       = "task_title"
	NoteColumnParentType = "parent_type"
)

type NotesModel struct {
	Note             sqlc.Note
	tableModel       table.Model
	totalWidth       int
	totalHeight      int
	horizontalMargin int
	verticalMargin   int
}

type SwitchToPreviousViewMsg struct{}

func (m *NotesModel) Init() tea.Cmd { return nil }

func (m *NotesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	default:
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// recalculateTable Recalculates the table based on the current margins
func (m *NotesModel) recalculateTable() {
	m.tableModel = m.tableModel.
		WithTargetWidth(m.calculateWidth()).
		WithMinimumHeight(m.calculateHeight())
}

func (m NotesModel) calculateWidth() int {
	return m.totalWidth - m.horizontalMargin
}

func (m NotesModel) calculateHeight() int {
	return m.totalHeight - m.verticalMargin - fixedVerticalMargin
}

func (m *NotesModel) View() string {
	body := strings.Builder{}

	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Add a new Note by pressing 'A'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Filter Archived Projects by pressing 'F'") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press left/right or page up/down to move between pages") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press 'space' to select a row, 'q' or 'ctrl+c' to quit") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press 'backspace' to delete row(s) after selecting or highlighting them.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)).Render("-Press 'ctrl+t' to switch to the Tasks View.") + "\n")
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Render("-Press 'ctrl+p' to switch to the Areas View.") + "\n")
	selectedIDs := []int64{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(int64))
	}

	// Convert selectedIDs to a slice of strings
	selectedIDStrings := make([]string, len(selectedIDs))
	for i, id := range selectedIDs {
		selectedIDStrings[i] = strconv.FormatInt(id, 10)
	}

	body.WriteString(
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(
				theme.Primary)).
			Render(
				fmt.Sprintf("Selected IDs: %s", strings.Join(selectedIDStrings, ", "))) + "\n")

	body.WriteString(m.tableModel.View())
	body.WriteString("\n")

	return body.String()
}

func NotesView() NotesModel {
	theme := tui.GetSelectedTheme()
	columns := []table.Column{
		table.NewColumn(NoteColumnKeyID, "ID", 5).WithStyle(
			lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color(theme.Secondary)).
				Align(lipgloss.Center)),
		table.NewColumn(NoteColumnKey, "Title", 15),
		table.NewFlexColumn(NoteColumnPath, "Path", 2),
		table.NewFlexColumn(NoteColumnLink, "Task", 1),
		table.NewFlexColumn(NoteColumnParentType, "Note Type", 1),
	}

	model := NotesModel{}
	var filteredRows []table.Row
	ctx := context.Background()
	conn, _, err := db.ConnectDB()
	if err != nil {
		log.Panicf("Notes View: There was an error connecting to the database: %v", err)
	}
	defer conn.Close()

	queries := sqlc.New(conn)
	allNotes, err := queries.ReadAllNotes(ctx)
	if err != nil {
		log.Fatalf("NewNotesModel: Error connecting to database: %v", err)
	}
	for _, note := range allNotes {
		newRow := table.NewRow(table.RowData{
			NoteColumnKeyID:      note.ID,
			NoteColumnKey:        note.Title,
			NoteColumnPath:       note.Path,
			NoteColumnLink:       note.AreaOrTaskTitle,
			NoteColumnParentType: note.ParentType,
		})
		filteredRows = append(filteredRows, newRow)
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
		SortByAsc(NoteColumnKeyID).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)),
			Data:  "<Missing Data>",
		})

	model.updateFooter()

	return model
}

func (m *NotesModel) refreshTableData() {
	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Printf("Error loading rows from database: %s", err)
	}

	m.tableModel = m.tableModel.WithRows(rows)

	m.updateFooter()
}

func (m *NotesModel) updateFooter() {
	highlightedRow := m.tableModel.HighlightedRow()
	rowID, ok := highlightedRow.Data[NoteColumnKeyID]
	if !ok {
		rowID = "No Rows Available"
	}

	footerText := fmt.Sprintf(
		"Pg. %d/%d - Currently looking at ID: %v",
		m.tableModel.CurrentPage(),
		m.tableModel.MaxPages(),
		rowID,
	)

	m.tableModel = m.tableModel.WithStaticFooter(footerText)
}

func (m *NotesModel) addNote() tea.Cmd {
	if m.tableModel.HighlightedRow().Data[NoteColumnKeyID] != nil {
		highlightedInfo := fmt.Sprintf("%v", m.tableModel.HighlightedRow().Data[NoteColumnKeyID])
		_, err := strconv.Atoi(highlightedInfo)
		if err != nil {
			log.Printf("Error converting ID to int: %s", err)
			return nil
		}

		form := &formInput.NewNoteForm{}
		theme := tui.GetSelectedTheme()
		err = form.NewNoteForm(*tui.ThemeGoTask(theme))
		if err != nil {
			log.Fatalf("Error creating form: %v", err)
		}

		if form.Submit {
			ctx := context.Background()
			conn, _, err := db.ConnectDB()
			if err != nil {
				log.Panicf("error connecting to database: %v", err)
			}
			queries := sqlc.New(conn)
			defer conn.Close()

			noteID, err := queries.GetNoteID(ctx)
			if err != nil && err != sql.ErrNoRows {
				log.Fatalf("Error getting note ID: %v", err)
			}

			err = queries.CreateNote(ctx, sqlc.CreateNoteParams{
				ID:    noteID,
				Title: form.Title,
				Path:  form.Path,
			},
			)
			if err != nil {
				log.Panicf("Error creating note: %s", err)
			}
			// FIXME: This is correctly adding to the TaskBridgeNotes/Notes tables now but when switching back to the datatable view, it needs to refresh the data.
			switch form.Type {
			case data.TaskNoteType:
				_, err = queries.CreateTaskBridgeNote(ctx, sqlc.CreateTaskBridgeNoteParams{
					NoteID:       noteID,
					ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
					ParentTaskID: sql.NullInt64{Int64: int64(form.ParentID), Valid: true},
				},
				)
				if err != nil {
					log.Panicf("Error creating task bridge note: %s", err)
					return nil
				}
			case data.AreaNoteType:
				_, err := queries.CreateAreaBridgeNote(ctx, sqlc.CreateAreaBridgeNoteParams{
					NoteID:       noteID,
					ParentCat:    sql.NullInt64{Int64: int64(data.AreaNoteType), Valid: true},
					ParentAreaID: sql.NullInt64{Int64: int64(form.ParentID), Valid: true},
				},
				)
				if err != nil {
					log.Panicf("Error creating area bridge note: %s", err)
				}

			}
			// Requery the database and update the table model
			rows, err := m.loadRowsFromDatabase()
			if err != nil {
				log.Printf("Error loading rows from database: %s", err)
				return nil
			}
			m.tableModel = m.tableModel.WithRows(rows)
			m.recalculateTable()

			// Update the footer
			m.updateFooter()
		}
	}

	return nil
}

func (m *NotesModel) loadRowsFromDatabase() ([]table.Row, error) {
	var filteredRows []table.Row
	ctx := context.Background()

	conn, _, err := db.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("loadRowsFromDatabase: error connecting to database: %w", err)
	}

	queries := sqlc.New(conn)
	defer conn.Close()
	notes, err := queries.ReadAllNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("loadRowsFromDatabase: error reading all notes: %w", err)
	}

	for _, note := range notes {
		newRow := table.NewRow(table.RowData{
			NoteColumnKeyID:      note.ID,
			NoteColumnKey:        note.Title,
			NoteColumnPath:       note.Path,
			NoteColumnLink:       note.AreaOrTaskTitle,
			NoteColumnParentType: note.ParentType,
		})
		filteredRows = append(filteredRows, newRow)
	}

	return filteredRows, nil
}

func (m *NotesModel) deleteNote() tea.Cmd {
	ctx := context.Background()
	selectedIDs := []int64{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(int64))
	}
	noteID := m.tableModel.HighlightedRow().Data[NoteColumnKeyID].(int64)

	conn, _, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
		return nil
	}
	defer conn.Close()
	queries := sqlc.New(conn)

	switch len(selectedIDs) {
	case 0:
		_, err := queries.DeleteNote(ctx, noteID)
		if err != nil {
			log.Fatalf("Error deleting task: %s", err)
			return nil
		}

	case 1:
		_, err := queries.DeleteNote(ctx, noteID)
		if err != nil {
			log.Fatalf("Error deleting task: %s", err)
			return nil
		}
	default:
		for _, noteID := range selectedIDs {
			_, err := queries.DeleteNote(ctx, noteID)
			if err != nil {
				log.Fatalf("Error deleting task: %s", err)
				return nil
			}
		}
	}

	rows, err := m.loadRowsFromDatabase()
	if err != nil {
		log.Fatalf("Error loading rows from database: %s", err)
		return nil
	}
	m.tableModel = m.tableModel.WithRows(rows)

	m.updateFooter()

	return nil
}

func (m *NotesModel) openNote() tea.Cmd {
	editor := config.GetEditorConfig()
	ctx := context.Background()
	selectedIDs := []int64{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(int64))
	}
	taskID := m.tableModel.HighlightedRow().Data[NoteColumnKeyID].(int64)

	conn, _, err := db.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %s", err)
		return nil
	}
	defer conn.Close()

	queries := sqlc.New(conn)
	if len(selectedIDs) == 1 {
		note, err := queries.ReadNoteByID(ctx, taskID)
		if err != nil {
			fmt.Printf("ReadNoteByID: There was an error reading the note: %v", err)
		}
		notePath, err := utils.ExpandPath(note.Path)

		utils.OpenNoteInEditor(editor, notePath)
		if err != nil {
			log.Fatalf("There was an error expanding the path: %v", err)
		}
	} else if len(selectedIDs) > 1 {
		notes, err := queries.ReadNoteByIDs(ctx, selectedIDs)
		if err != nil {
			fmt.Printf("ReadNoteByID: There was an error reading the note: %v", err)
		}
		notePaths := make([]string, len(notes))
		for i, note := range notes {
			notePath, err := utils.ExpandPath(note.Path)
			if err != nil {
				log.Fatalf("There was an error expanding the path: %v", err)
			}
			notePaths[i] = notePath
		}
		utils.OpenNoteInEditor(editor, notePaths...)

	} else if len(selectedIDs) == 0 {

		note, err := queries.ReadNoteByID(ctx, taskID)
		if err != nil {
			fmt.Printf("ReadNoteByID: There was an error reading the note: %v", err)
		}
		notePath, err := utils.ExpandPath(note.Path)

		utils.OpenNoteInEditor(editor, notePath)
		if err != nil {
			log.Fatalf("There was an error expanding the path: %v", err)
		}
	}

	// Update the footer
	m.updateFooter()

	return nil
}
