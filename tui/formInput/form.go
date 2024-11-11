package formInput

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

// TODO: In the future make this pull from config/default color theme for app?
type ColorTheme struct {
	BorderForeground string
	BorderBackground string
}

func colorTheme() ColorTheme {
	return ColorTheme{
		BorderForeground: "#ea9a97",
		BorderBackground: "#232136",
	}
}

func DefaultStyles() *Styles {
	colors := colorTheme()
	s := new(Styles)
	s.BorderColor = lipgloss.Color(colors.BorderForeground)
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(80)
	return s
}

type Model struct {
	styles    *Styles
	index     int
	questions []Question
	width     int
	height    int
	done      bool
}

type Question struct {
	question string
	answer   string
	input    Input
}

func NewQuestion(q string) Question {
	return Question{question: q}
}

func newShortQuestion(q string) Question {
	question := NewQuestion(q)
	model := NewShortAnswerField()
	question.input = model
	return question
}

func newLongQuestion(q string) Question {
	question := NewQuestion(q)
	model := NewLongAnswerField()
	question.input = model
	return question
}

func New(questions []Question) *Model {
	styles := DefaultStyles()

	return &Model{
		styles:    styles,
		questions: questions,
	}
}

func (m Model) Init() tea.Cmd {
	return m.questions[m.index].input.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	current := &m.questions[m.index]
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.index == len(m.questions)-1 {
				m.done = true
			}
			current.answer = current.input.Value()
			m.Next()
			return m, current.input.Blur
		}
	}
	current.input, cmd = current.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	current := m.questions[m.index]
	if m.done {
		var output string
		for _, q := range m.questions {
			output += fmt.Sprintf("%s: %s\n", q.question, q.answer)
		}
		return output
	}
	if m.width == 0 {
		return "loading..."
	}
	// stack some left-aligned strings together in the center of the window
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			current.question,
			m.styles.InputField.Render(current.input.View()),
		),
	)
}

func (m *Model) Next() {
	if m.index < len(m.questions)-1 {
		m.index++
	} else {
		m.index = 0
	}
}

func Run() {
	questions := []Question{newShortQuestion("What is your task?"), newShortQuestion("Priority Level?"), newShortQuestion("Status?"), newShortQuestion("Notes?")}
	model := New(questions)
	RunModel(model)
}

func RunModel(m *Model) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
}
