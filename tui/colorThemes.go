package tui

type rosepine struct {
	Overlay string
	Love    string
	Gold    string
	Rose    string
	Pine    string
	Foam    string
	Iris    string
}

type rosepinemoon struct {
	Overlay string
	Love    string
	Gold    string
	Rose    string
	Pine    string
	Foam    string
	Iris    string
}

type rosepinedawn struct {
	Overlay string
	Love    string
	Gold    string
	Rose    string
	Pine    string
	Foam    string
	Iris    string
}

type colorThemes struct {
	RosePine     rosepine
	RosePineMoon rosepinemoon
	RosePineDawn rosepinedawn
}

var Themes = colorThemes{
	RosePine: rosepine{
		Overlay: "#26233a",
		Love:    "#eb6f92",
		Gold:    "#f6c177",
		Rose:    "#ebbcba",
		Pine:    "#31748f",
		Foam:    "#9ccfd8",
		Iris:    "#c4a7e7",
	},
	RosePineMoon: rosepinemoon{
		Overlay: "#393552",
		Love:    "#eb6f92",
		Gold:    "#f6c177",
		Rose:    "#ea9a97",
		Pine:    "#3e8fb0",
		Foam:    "#9ccfd8",
		Iris:    "#c4a7e7",
	},
	RosePineDawn: rosepinedawn{
		Overlay: "#f2e9e1",
		Love:    "#b4637a",
		Gold:    "#ea9d34",
		Rose:    "#d7827e",
		Pine:    "#286983",
		Foam:    "#56949f",
		Iris:    "#907aa9",
	},
}
