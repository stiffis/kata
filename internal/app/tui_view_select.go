package app

import (
	"fmt"
	"strings"

	"kata/pkg/themes"
)

func (m model) renderThemeSelect() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("🎨 Select Theme"))
	b.WriteString("\n\n")

	themeNames := themes.ListThemes()

	// Show preview of selected theme
	if m.themeIndex < len(themeNames) {
		selectedTheme := themes.GetTheme(themeNames[m.themeIndex])
		b.WriteString(m.theme.Dim.Render("Preview:"))
		b.WriteString("\n")
		b.WriteString(selectedTheme.Title.Render("🥋 KATA"))
		b.WriteString("  ")
		b.WriteString(selectedTheme.Dim.Render("\"Practice makes perfect\""))
		b.WriteString("\n")

		// Simulate typing practice
		sampleText := "func main() {"
		userText := "func mai"
		for i, ch := range sampleText {
			if i < len(userText) {
				if i < len(userText) && userText[i] == byte(ch) {
					b.WriteString(selectedTheme.Correct.Render(string(ch)))
				} else {
					b.WriteString(selectedTheme.Incorrect.Render(string(ch)))
				}
			} else if i == len(userText) {
				b.WriteString(selectedTheme.Cursor.Render(string(ch)))
			} else {
				b.WriteString(selectedTheme.Dim.Render(string(ch)))
			}
		}
		b.WriteString("\n")
		b.WriteString(selectedTheme.Stats.Render("WPM: 45 | Accuracy: 92%"))
		b.WriteString("\n")
		b.WriteString(selectedTheme.Separator.Render(strings.Repeat("─", 40)))
		b.WriteString("\n\n")
	}

	// Theme list
	b.WriteString(m.theme.Menu.Render("Available Themes:"))
	b.WriteString("\n")
	for i, themeName := range themeNames {
		cursor := "  "
		style := m.theme.Menu
		if i == m.themeIndex {
			cursor = "▶ "
			style = m.theme.Selected
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%-12s", cursor, themeName)))

		// Mini color indicators
		preview := themes.GetTheme(themeName)
		b.WriteString("  ")
		b.WriteString(preview.Correct.Render("●"))
		b.WriteString(preview.Incorrect.Render("●"))
		b.WriteString(preview.Stats.Render("●"))
		b.WriteString(preview.Menu.Render("●"))
		b.WriteString(preview.Title.Render("●"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("↑/↓ or j/k to navigate | Enter to apply | ESC to cancel"))

	return b.String()
}

func (m model) renderLanguageSelect() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("🌍 Select Language"))
	b.WriteString("\n\n")
	b.WriteString(m.theme.Dim.Render("Choose the vocabulary for your practice:"))
	b.WriteString("\n\n")

	languages := []string{"go", "cpp", "javascript", "rust", "python", "english", "spanish", "french", "german"}

	for i, lang := range languages {
		cursor := "  "
		style := m.theme.Menu
		if i == m.themeIndex {
			cursor = "▶ "
			style = m.theme.Selected
		}

		// Add active indicator
		active := ""
		if lang == m.config.Language {
			active = " " + m.theme.Correct.Render("(current)")
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%-10s%s", cursor, strings.Title(lang), active)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("↑/↓ or j/k to navigate | Enter to apply | ESC to cancel"))

	return b.String()
}
