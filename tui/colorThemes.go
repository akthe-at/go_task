package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func GetSelectedTheme() Theme {
	switch Themes.Selected.Theme {
	case "rose-pine":
		return Themes.RosePine
	case "rose-pine-moon":
		return Themes.RosePineMoon
	case "rose-pine-dawn":
		return Themes.RosePineDawn
	default:
		if Theme, exists := Themes.Additional[Themes.Selected.Theme]; exists {
			return Theme
		}
		return Themes.RosePine // default theme
	}
}

type Theme struct {
	Primary    string `toml:"primary"`
	Secondary  string `toml:"secondary"`
	Background string `toml:"background"`
	Foreground string `toml:"foreground"`
	Success    string `toml:"success"`
	Warning    string `toml:"warning"`
	Accent     string `toml:"accent"`
}

type SelectedTheme struct {
	Theme string `toml:"theme"`
}

type ColorThemes struct {
	RosePine     Theme            `toml:"rose_pine"`
	RosePineMoon Theme            `toml:"rose_pine_moon"`
	RosePineDawn Theme            `toml:"rose_pine_dawn"`
	Additional   map[string]Theme `toml:"additional"`
	Selected     SelectedTheme    `toml:"selected"`
}

var Themes = ColorThemes{
	RosePine: Theme{
		Primary:    "#eb6f92",
		Secondary:  "#31748f",
		Background: "#26233a",
		Foreground: "#9ccfd8",
		Success:    "#f6c177",
		Warning:    "#ebbcba",
		Accent:     "#c4a7e7",
	},
	RosePineMoon: Theme{
		Primary:    "#eb6f92",
		Secondary:  "#3e8fb0",
		Background: "#393552",
		Foreground: "#9ccfd8",
		Success:    "#f6c177",
		Warning:    "#ea9a97",
		Accent:     "#c4a7e7",
	},
	RosePineDawn: Theme{
		Primary:    "#ea9d34",
		Accent:     "#b4637a",
		Background: "#56949f",
		Foreground: "#286983",
		Success:    "#d7827e",
		Secondary:  "#907aa9",
		Warning:    "#f2e9e1",
	},
	Additional: make(map[string]Theme),
	Selected:   SelectedTheme{Theme: "RosePine"},
}

func ThemeGoTask(theme Theme) *huh.Theme {
	t := huh.ThemeBase()
	var (
		primary    = theme.Primary
		secondary  = theme.Secondary
		background = theme.Background
		foreground = theme.Foreground
		success    = theme.Success
		warning    = theme.Warning
		accent     = theme.Accent
	)

	t.Focused.Base = t.Focused.Base.BorderForeground(lipgloss.Color(foreground))
	t.Focused.Title = t.Focused.Title.Foreground(lipgloss.Color(primary))
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(lipgloss.Color(foreground))
	t.Focused.Description = t.Focused.Description.Foreground(lipgloss.Color(accent))
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(lipgloss.Color(warning))
	t.Focused.Directory = t.Focused.Directory.Foreground(lipgloss.Color(accent))
	t.Focused.File = t.Focused.File.Foreground(lipgloss.Color(success))
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(lipgloss.Color(warning))
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(lipgloss.Color(accent))
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(lipgloss.Color(primary))
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(lipgloss.Color(secondary))
	t.Focused.Option = t.Focused.Option.Foreground(lipgloss.Color(foreground))
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(lipgloss.Color(accent))
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(lipgloss.Color(success))
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(lipgloss.Color(foreground))
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(lipgloss.Color(foreground))
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(lipgloss.Color(secondary))
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color(accent)).Background(lipgloss.Color(primary)).Bold(true)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(lipgloss.Color(foreground)).Background(lipgloss.Color(background))
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(lipgloss.Color(accent))
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(lipgloss.Color(secondary))
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(lipgloss.Color(secondary))

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	return t
}
