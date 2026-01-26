package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) renderMenu() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("🥋 KATA - The Way of the Keyboard"))
	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("\"Slow is smooth. Smooth is fast.\""))
	b.WriteString("\n\n")

	for i, option := range m.menuOptions {
		cursor := "  "
		style := m.theme.Menu
		if i == m.menuIndex {
			cursor = "▶ "
			style = m.theme.Selected
		}

		// Add indicator for Toggle Zen Mode
		displayOption := option
		if option == "Toggle Zen Mode" {
			if m.config.ZenMode {
				displayOption = "Toggle Zen Mode [ON]"
			} else {
				displayOption = "Toggle Zen Mode [OFF]"
			}
		}

		b.WriteString(style.Render(cursor + displayOption))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("↑/↓ or j/k to navigate | Enter to select | q to quit"))

	content := b.String()
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}
