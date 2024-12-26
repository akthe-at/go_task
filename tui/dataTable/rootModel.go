package datatable

import (
	"github.com/akthe-at/go_task/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	NotesTableView View = iota
	TasksTableView
	AreasTableView
)

var theme = tui.GetSelectedTheme()

type (
	View                         int
	AddNoteMsg                   struct{}
	AddTaskMsg                   struct{}
	AddAreaMsg                   struct{}
	SwitchToTasksTableViewMsg    struct{}
	SwitchToProjectsTableViewMsg struct{}
)

type RootModel struct {
	Height int
	Width  int

	Tasks TaskModel
	Notes NotesModel
	Areas AreasModel

	CurrentView  View
	PreviousView View
}

func NewRootModel() RootModel {
	return RootModel{
		Tasks:       TaskViewModel(),
		Notes:       NotesView(),
		Areas:       AreaViewModel(),
		CurrentView: TasksTableView,
	}
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) isInitialized() bool {
	return m.Height != 0 && m.Width != 0
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.isInitialized() {
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// case "enter":
		// 	switch m.CurrentView {
		// 	case TasksTableView:
		// 		highlightedRow := m.Tasks.tableModel.HighlightedRow()
		// 		m.Tasks.PrintRow(highlightedRow)
		// 	case NotesTableView:
		// 		m.Notes.recalculateTable()
		// 	case ProjectsTableView:
		// 		m.Projects.recalculateTable()
		// 	}
		case "ctrl+c":
			return m, tea.Quit
		case "D":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.deleteTask()
			case NotesTableView:
				m.Notes.deleteNote()
			case AreasTableView:
				m.Areas.deleteArea()
			}
		case "O":
			switch m.CurrentView {
			case NotesTableView:
				m.Notes.openNote()
			}
		case "A":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.addTask()
			case NotesTableView:
				m.Notes.addNote()
			case AreasTableView:
				m.Areas.addArea()
			}
		case "T":
			switch m.CurrentView {
			case AreasTableView:
				m.Areas.addTaskToArea()
			case TasksTableView:
				m.Tasks.recalculateTable()
			case NotesTableView:
				m.Notes.recalculateTable()
			}
		case "ctrl+t":
			m.Tasks.refreshTableData()
			m.PreviousView = m.CurrentView
			m.CurrentView = TasksTableView
		case "ctrl+T":
			m.Task.refreshTableData()
			m.PreviousView = m.CurrentView
			m.CurrentView = TaskTableView
		case "ctrl+n":
			m.Notes.refreshTableData()
			m.PreviousView = m.CurrentView
			m.CurrentView = NotesTableView
		case "ctrl+p":
			m.PreviousView = m.CurrentView
			m.CurrentView = AreasTableView
		}

	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
		msg.Height -= 2
		msg.Width -= 4
		return m.propagate(msg), nil
	case SwitchToTasksTableViewMsg:
		m.CurrentView = TasksTableView
	case SwitchToTaskTableViewMsg:
		m.CurrentView = TaskTableView
	case SwitchToProjectsTableViewMsg:
		m.CurrentView = AreasTableView
	case SwitchToPreviousViewMsg:
		m.CurrentView = m.PreviousView
	case AddNoteMsg:
		updatedNotes, _ := m.Notes.Update(msg)
		m.Notes = *updatedNotes.(*NotesModel)
	case AddAreaMsg:
		updatedAreas, _ := m.Areas.Update(msg)
		m.Areas = *updatedAreas.(*AreasModel)
	case AddTaskMsg:
		updatedTasks, _ := m.Tasks.Update(msg)
		m.Tasks = *updatedTasks.(*TaskModel)
	}

	return m.propagate(msg), nil
}

func (m *RootModel) propagate(msg tea.Msg) tea.Model {
	var updatedTasks tea.Model
	var updatedNotes tea.Model
	var updatedProjects tea.Model

	updatedTasks, _ = m.Tasks.Update(msg)
	m.Tasks = *updatedTasks.(*TaskModel)

	updatedNotes, _ = m.Notes.Update(msg)
	m.Notes = *updatedNotes.(*NotesModel)

	updatedProjects, _ = m.Areas.Update(msg)
	m.Areas = *updatedProjects.(*AreasModel)

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		msg.Height -= m.Notes.totalHeight
		msg.Width -= m.Notes.totalWidth
		return m
	}
	return m
}

func (m RootModel) View() string {
	var s lipgloss.Style
	switch m.CurrentView {
	case TasksTableView:
		return s.Render(m.Tasks.View())
	case NotesTableView:
		return s.Render(m.Notes.View())
	case AreasTableView:
		return s.Render(m.Areas.View())
	default:
		return s.Render("")
	}
}
