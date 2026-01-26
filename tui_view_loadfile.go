package main

import (
	"strings"
)

func (m model) renderLoadFile() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("📂 Load File"))
	b.WriteString("\n\n")
	b.WriteString(m.theme.Dim.Render("Enter the path to a text file:"))
	b.WriteString("\n\n")

	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.errMsg != "" {
		b.WriteString(m.theme.Incorrect.Render(m.errMsg))
		b.WriteString("\n\n")
	}

	b.WriteString(m.theme.Dim.Render("Enter to load | ESC to cancel"))

	return b.String()
}
