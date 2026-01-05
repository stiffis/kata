package themes

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
	Correct    lipgloss.Style
	Incorrect  lipgloss.Style
	Cursor     lipgloss.Style
	Dim        lipgloss.Style
	Title      lipgloss.Style
	Stats      lipgloss.Style
	Menu       lipgloss.Style
	Selected   lipgloss.Style
	Separator  lipgloss.Style
}

var availableThemes = map[string]Theme{
	"default": DefaultTheme(),
	"catppuccin": CatppuccinTheme(),
	"rose-pine": RosePineTheme(),
	"dracula": DraculaTheme(),
	"nord": NordTheme(),
	"gruvbox": GruvboxTheme(),
}

func DefaultTheme() Theme {
	return Theme{
		Name:      "default",
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Underline(true),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")),
		Stats:     lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		Menu:      lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		Selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}

func CatppuccinTheme() Theme {
	return Theme{
		Name:      "catppuccin",
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")), // Green
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")), // Red
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f5e0dc")).Underline(true), // Rosewater
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")), // Overlay0
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#cba6f7")), // Mauve
		Stats:     lipgloss.NewStyle().Foreground(lipgloss.Color("#89dceb")), // Sky
		Menu:      lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")), // Blue
		Selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true), // Green
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a")), // Surface1
	}
}

func RosePineTheme() Theme {
	return Theme{
		Name:      "rose-pine",
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9ccfd8")), // Foam
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("#eb6f92")), // Love
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f6c177")).Underline(true), // Gold
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6e6a86")), // Muted
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#c4a7e7")), // Iris
		Stats:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ebbcba")), // Rose
		Menu:      lipgloss.NewStyle().Foreground(lipgloss.Color("#31748f")), // Pine
		Selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("#9ccfd8")).Bold(true), // Foam
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("#26233a")), // Surface
	}
}

func DraculaTheme() Theme {
	return Theme{
		Name:      "dracula",
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")), // Green
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")), // Red
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f1fa8c")).Underline(true), // Yellow
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")), // Comment
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#bd93f9")), // Purple
		Stats:     lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")), // Cyan
		Menu:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6")), // Pink
		Selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true), // Green
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("#44475a")), // Current Line
	}
}

func NordTheme() Theme {
	return Theme{
		Name:      "nord",
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a3be8c")), // Green
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("#bf616a")), // Red
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ebcb8b")).Underline(true), // Yellow
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#4c566a")), // Polar Night
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b48ead")), // Purple
		Stats:     lipgloss.NewStyle().Foreground(lipgloss.Color("#88c0d0")), // Frost
		Menu:      lipgloss.NewStyle().Foreground(lipgloss.Color("#81a1c1")), // Blue
		Selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("#a3be8c")).Bold(true), // Green
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("#3b4252")), // Dark
	}
}

func GruvboxTheme() Theme {
	return Theme{
		Name:      "gruvbox",
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("#b8bb26")), // Green
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("#fb4934")), // Red
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#fabd2f")).Underline(true), // Yellow
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")), // Gray
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d3869b")), // Purple
		Stats:     lipgloss.NewStyle().Foreground(lipgloss.Color("#8ec07c")), // Aqua
		Menu:      lipgloss.NewStyle().Foreground(lipgloss.Color("#83a598")), // Blue
		Selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("#b8bb26")).Bold(true), // Green
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("#504945")), // Dark
	}
}

func GetTheme(name string) Theme {
	if theme, ok := availableThemes[name]; ok {
		return theme
	}
	return DefaultTheme()
}

func ListThemes() []string {
	themes := []string{
		"default",
		"catppuccin",
		"rose-pine",
		"dracula",
		"nord",
		"gruvbox",
	}
	return themes
}
