package datatable

import (
	"github.com/akthe-at/go_task/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	NotesTableView View = iota
	TasksTableView
	ProjectsTableView
)

var theme = tui.GetSelectedTheme()

type (
	View                         int
	AddNoteMsg                   struct{}
	AddTaskMsg                   struct{}
	AddProjectMsg                struct{}
	SwitchToTasksTableViewMsg    struct{}
	SwitchToProjectsTableViewMsg struct{}
)

type RootModel struct {
	Height int
	Width  int

	Tasks    TaskModel
	Notes    NotesModel
	Projects AreasModel

	CurrentView  View
	PreviousView View
}

func NewRootModel() RootModel {
	return RootModel{
		Tasks:       TaskViewModel(),
		Notes:       NotesView(),
		Projects:    ProjectViewModel(),
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
			if m.CurrentView == TasksTableView {
				m.Tasks.deleteTask()
			} else {
				m.Notes.deleteNote()
			}
		case "A":
			switch m.CurrentView {
			case TasksTableView:
				m.Tasks.addTask()
			case NotesTableView:
				m.Notes.addNote()
			case ProjectsTableView:
				m.Projects.addArea()
			}
		case "T":
			switch m.CurrentView {
			case ProjectsTableView:
				m.Projects.addTaskToArea()
			case TasksTableView:
				m.Tasks.recalculateTable()
			case NotesTableView:
				m.Notes.recalculateTable()
			}
		case "ctrl+t":
			// tui.ClearTerminalScreen()
			m.Tasks.refreshTableData()
			m.PreviousView = m.CurrentView
			m.CurrentView = TasksTableView
		case "ctrl+n":
			// tui.ClearTerminalScreen()
			m.Notes.refreshTableData()
			m.PreviousView = m.CurrentView
			m.CurrentView = NotesTableView
		case "ctrl+p":
			// tui.ClearTerminalScreen()
			m.PreviousView = m.CurrentView
			m.CurrentView = ProjectsTableView
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
		m.CurrentView = ProjectsTableView
	case SwitchToPreviousViewMsg:
		m.CurrentView = m.PreviousView
	case AddNoteMsg:
		updatedNotes, _ := m.Notes.Update(msg)
		m.Notes = *updatedNotes.(*NotesModel)
	case AddProjectMsg:
		updatedProjects, _ := m.Projects.Update(msg)
		m.Projects = *updatedProjects.(*AreasModel)
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

	updatedProjects, _ = m.Projects.Update(msg)
	m.Projects = *updatedProjects.(*AreasModel)

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
	case ProjectsTableView:
		return s.Render(m.Projects.View())
	default:
		return s.Render("")
	}
}
