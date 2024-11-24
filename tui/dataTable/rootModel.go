package datatable

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	NotesTableView View = iota
	TasksTableView
)


type (
	View                      int
	AddNoteMsg                struct{}
	AddTaskMsg                struct{}
	SwitchToTasksTableViewMsg struct{}
)

type RootModel struct {
	Height int
	Width  int

	Tasks TaskModel
	Notes NotesModel

	CurrentView  View
	PreviousView View
}

func NewRootModel() RootModel {
	return RootModel{
		Tasks:       TaskViewModel(),
		Notes:       NotesView(),
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
		case "D":
			if m.CurrentView == TasksTableView {
				m.Tasks.deleteTask()
			} else {
				m.Notes.deleteNote()
			}
		case "ctrl+t":
			m.PreviousView = m.CurrentView
			m.CurrentView = TasksTableView
		case "ctrl+n":
			m.PreviousView = m.CurrentView
			m.CurrentView = NotesTableView
		}

	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
		msg.Height -= 2
		msg.Width -= 4
		return m.propagate(msg), nil
	case SwitchToTasksTableViewMsg:
		m.CurrentView = TasksTableView
	case SwitchToPreviousViewMsg:
		m.CurrentView = m.PreviousView
	case AddNoteMsg:
		updatedNotes, _ := m.Notes.Update(msg)
		m.Notes = *updatedNotes.(*NotesModel)
	case AddTaskMsg:
		updatedTasks, _ := m.Tasks.Update(msg)
		m.Tasks = *updatedTasks.(*TaskModel)
	}

	return m.propagate(msg), nil
}

func (m *RootModel) propagate(msg tea.Msg) tea.Model {
	var updatedTasks tea.Model
	var updatedNotes tea.Model

	updatedTasks, _ = m.Tasks.Update(msg)
	m.Tasks = *updatedTasks.(*TaskModel)

	updatedNotes, _ = m.Notes.Update(msg)
	m.Notes = *updatedNotes.(*NotesModel)

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
	default:
		return s.Render("")
	}
}
