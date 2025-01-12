package datatable

import (
	"github.com/akthe-at/go_task/data"
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
		case "ctrl+c":
			return m, tea.Quit
		case "backspace":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.deleteTask()
			case NotesTableView:
				m.Notes.deleteNote()
			case AreasTableView:
				m.Areas.deleteArea()
			}
		case "p":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.updateStatus(data.StatusPlanning)
			case AreasTableView:
				m.Areas.updateStatus(data.StatusPlanning)
			}
		case "P":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.togglePriorityStatus()
				// case NotesTableView:
				// m.Areas.updateStatus(data.StatusDone)
			}
		case "a":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.archiveTask()
			case AreasTableView:
				m.Areas.archiveArea()
			}
		case "d":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.updateStatus(data.StatusDoing)
			case AreasTableView:
				m.Areas.updateStatus(data.StatusDoing)
			}
		case "t":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.updateStatus(data.StatusToDo)
			case AreasTableView:
				m.Areas.updateStatus(data.StatusToDo)
			}
		case "D":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.updateStatus(data.StatusDone)
			case AreasTableView:
				m.Areas.updateStatus(data.StatusDone)
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
		case "ctrl+n":
			m.Notes.refreshTableData()
			m.PreviousView = m.CurrentView
			m.CurrentView = NotesTableView
		case "ctrl+p":
			m.Areas.refreshTableData()
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
	var updatedAreas tea.Model

	updatedTasks, _ = m.Tasks.Update(msg)
	m.Tasks = *updatedTasks.(*TaskModel)

	updatedNotes, _ = m.Notes.Update(msg)
	m.Notes = *updatedNotes.(*NotesModel)

	updatedAreas, _ = m.Areas.Update(msg)
	m.Areas = *updatedAreas.(*AreasModel)

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
