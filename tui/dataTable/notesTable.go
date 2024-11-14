package datatable

import (
	"fmt"
	"log"
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
	NoteColumnKeyID    = "id"
	NoteColumnKey      = "title"
	NoteColumnPath     = "path"
	NoteColumnLink     = "task_title"
	NoteColumnParentID = "parent_id"
)

type NotesModel struct {
	tableModel       table.Model
	Note             data.Note
	totalWidth       int
	totalHeight      int
	horizontalMargin int
	verticalMargin   int
}

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
		case "A":
			cmds = append(cmds, m.addNote())
		case "ctrl+t":
			taskView := TaskViewModel()
			return taskView.Update(msg)
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
	}

	return m, tea.Batch(cmds...)
}

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

	selectedIDs := []string{}

	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(string))
	}

	body.WriteString(
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(
				tui.Themes.RosePineMoon.Love)).
			Render(
				fmt.Sprintf("Selected IDs: %s", strings.Join(selectedIDs, ", "))) + "\n")

	body.WriteString(m.tableModel.View())
	body.WriteString("\n")

	return body.String()
}

func NotesView() NotesModel {
	columns := []table.Column{
		table.NewColumn(NoteColumnKeyID, "ID", 5).WithStyle(
			lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Pine)).
				Align(lipgloss.Center)),
		table.NewFlexColumn(NoteColumnKey, "Title", 1),
		table.NewFlexColumn(NoteColumnPath, "Path", 3),
		table.NewFlexColumn(NoteColumnLink, "Task", 1),
	}

	model := NotesModel{}
	var filteredRows []table.Row
	// I think I could embed a Notes struct into the NotesModel struct, then use that to query some data?

	conn, err := db.ConnectDB()
	if err != nil {
		panic("")
	}
	defer conn.Close()
	notes, err := model.Note.ReadAll(conn, data.TaskNoteType)
	if err != nil {
		log.Fatalf("NewNotesModel: Error connecting to database: %v", err)
	}
	for _, note := range notes {
		newRow := table.NewRow(table.RowData{
			NoteColumnKeyID:    note.NoteID,
			NoteColumnKey:      note.NoteTitle,
			NoteColumnPath:     note.NotePath,
			NoteColumnLink:     note.LinkTitle,
			NoteColumnParentID: note.ParentID,
		})
		filteredRows = append(filteredRows, newRow)
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
		SortByAsc(NoteColumnKeyID).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color(tui.Themes.RosePineMoon.Love)),
			Data:  "<Missing Data>",
		})

	model.updateFooter()

	return model
}

func (m *NotesModel) updateFooter() {
	highlightedRow := m.tableModel.HighlightedRow()
	rowID, ok := highlightedRow.Data[NoteColumnKeyID]
	if !ok {
		rowID = "No Rows Available"
	}

	footerText := fmt.Sprintf(
		"Pg. %d/%d - Currently looking at ID: %d",
		m.tableModel.CurrentPage(),
		m.tableModel.MaxPages(),
		rowID,
	)

	m.tableModel = m.tableModel.WithStaticFooter(footerText)
}

func (m *NotesModel) addNote() tea.Cmd {
	selectedIDs := []string{}

	// FIXME: HERE
	//
	for _, row := range m.tableModel.SelectedRows() {
		selectedIDs = append(selectedIDs, row.Data[NoteColumnKeyID].(string))
	}
	highlightedInfo := fmt.Sprintf("%v", m.tableModel.HighlightedRow().Data[NoteColumnKeyID])
	taskID, err := strconv.Atoi(highlightedInfo)
	if err != nil {
		log.Printf("Error converting ID to int: %s", err)
		return nil
	}

	form := &formInput.NewNoteForm{}
	err = form.NewForm()
	if err != nil {
		log.Fatalf("Error creating form: %v", err)
	}

	if form.Submit {

		newNote := data.Note{
			Title: form.Title,
			Path:  form.Path,
		}

		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		note := data.Note{
			ID:    newNote.ID,
			Path:  newNote.Path,
			Title: newNote.Title,
			Type:  newNote.Type,
		}
		err = note.Create(conn, taskID)
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

func (m *NotesModel) loadRowsFromDatabase() ([]table.Row, error) {
	var filteredRows []table.Row
	// I think I could embed a Notes struct into the NotesModel struct, then use that to query some data?

	conn, err := db.ConnectDB()
	if err != nil {
		panic("")
	}
	defer conn.Close()
	notes, err := m.Note.ReadAll(conn, data.TaskNoteType)
	if err != nil {
		log.Fatalf("NewNotesModel: Error connecting to database: %v", err)
	}
	for _, note := range notes {
		newRow := table.NewRow(table.RowData{
			NoteColumnKeyID:    note.NoteID,
			NoteColumnKey:      note.NoteTitle,
			NoteColumnPath:     note.NotePath,
			NoteColumnLink:     note.LinkTitle,
			NoteColumnParentID: note.ParentID,
		})
		filteredRows = append(filteredRows, newRow)
	}

	return filteredRows, nil
}

// func RunModel(m *Model) {
// 	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
// 		fmt.Println("Error running program:", err)
// 		os.Exit(1)
// 	}
// }

//
// // func (m *Model) loadRowsFromDatabase() ([]table.Row, error) {
// // 	conn, err := db.ConnectDB()
// // 	if err != nil {
// // 		return nil, fmt.Errorf("error connecting to database: %w", err)
// // 	}
// // 	defer conn.Close()
// //
// // 	task := data.Task{}
// // 	tasks, err := task.ReadAll(conn)
// // 	if err != nil {
// // 		return nil, fmt.Errorf("error reading tasks: %w", err)
// // 	}
// //
// // 	var rows []table.Row
// // 	for _, task := range tasks {
// // 		row := table.NewRow(table.RowData{
// // 			columnKeyID:        fmt.Sprintf("%d", task.ID),
// // 			columnKeyTask:      task.Title,
// // 			columnKeyPriority:  task.Priority,
// // 			columnKeyStatus:    task.Status,
// // 			columnKeyArchived:  fmt.Sprintf("%t", task.Archived),
// // 			columnKeyCreatedAt: task.CreatedAt.Format("2006-01-02 15:04:05"),
// // 			columnKeyDueDate:   task.DueDate.Format("2006-01-02 15:04:05"),
// // 			columnKeyNotes:     task.NoteTitles, // extractNoteTitles(task.Notes),
// // 		})
// // 		rows = append(rows, row)
// // 	}
// //
// // 	return rows, nil
// // }
