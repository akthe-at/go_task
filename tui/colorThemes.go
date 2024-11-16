package tui

func GetSelectedTheme() Theme {
	switch Themes.Selected.Theme {
	case "rose-pine":
		return Themes.RosePine
	case "rose-pine-moon":
		return Themes.RosePineMoon
	case "rose-pine-dawn":
		return Themes.RosePineDawn
	default:
		if theme, exists := Themes.Additional[Themes.Selected.Theme]; exists {
			return theme
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
