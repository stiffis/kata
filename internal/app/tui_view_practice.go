package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) renderPractice() string {
	var content string

	// Use Zen mode if enabled
	if m.config.ZenMode {
		content = m.renderPracticeZen()
	} else {
		var b strings.Builder

		b.WriteString(m.theme.Title.Render("🥋 KATA - The Way of the Keyboard"))
		b.WriteString("\n\n")

		if m.engine.IsFinished {
			wpm, accuracy, duration := m.engine.GetStats()

			b.WriteString(m.theme.Stats.Render(fmt.Sprintf("✓ Complete!\n\n")))
			b.WriteString(m.theme.Stats.Render(fmt.Sprintf("Time: %.1f seconds\n", duration)))
			b.WriteString(m.theme.Stats.Render(fmt.Sprintf("WPM: %.0f\n", wpm)))
			b.WriteString(m.theme.Stats.Render(fmt.Sprintf("Accuracy: %.1f%%\n", accuracy)))
			b.WriteString("\n")
			b.WriteString(m.theme.Dim.Render("Press Enter to return to menu | q to quit"))

			content = b.String()
		} else {
			// Render the text with colors
			// Wrap text in a container to prevent it from spreading too wide
			var textBlock strings.Builder

			targetText := m.engine.TargetText
			userInput := m.engine.UserInput

			for i := 0; i < len(targetText); i++ {
				if i < len(userInput) {
					if userInput[i] == targetText[i] {
						textBlock.WriteString(m.theme.Correct.Render(string(targetText[i])))
					} else {
						textBlock.WriteString(m.theme.Incorrect.Render(string(userInput[i])))
					}
				} else if i == len(userInput) {
					textBlock.WriteString(m.theme.Cursor.Render(string(targetText[i])))
				} else {
					textBlock.WriteString(m.theme.Dim.Render(string(targetText[i])))
				}
			}

			// Show cursor if user typed past the end
			if len(userInput) >= len(targetText) {
				textBlock.WriteString(m.theme.Cursor.Render(" "))
			}

			// Apply a width limit to the text block for better reading
			// Default width 60, but adapt if screen is smaller
			textWidth := 60
			if m.width > 0 && m.width < 70 {
				textWidth = m.width - 10
			}
			if textWidth < 20 {
				textWidth = 20
			}

			style := lipgloss.NewStyle().Width(textWidth).Align(lipgloss.Left)
			b.WriteString(style.Render(textBlock.String()))

			b.WriteString("\n\n")

			if !m.engine.StartTime.IsZero() {
				wpm, _, duration := m.engine.GetStats()
				progress := float64(len(userInput)) / float64(len(targetText)) * 100.0
				if progress > 100.0 {
					progress = 100.0
				}
				b.WriteString(m.theme.Dim.Render(fmt.Sprintf("Progress: %.0f%% | Time: %.0fs | Errors: %d | WPM: %.0f", progress, duration, m.engine.ErrorCount, wpm)))
			} else {
				b.WriteString(m.theme.Dim.Render("Start typing to begin..."))
			}

			b.WriteString("\n\n")
			b.WriteString(m.theme.Dim.Render("ESC to menu | Ctrl+Z to toggle zen | Ctrl+C to quit"))

			content = b.String()
		}
	}

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m model) renderPracticeZen() string {
	var b strings.Builder

	// Zen Mode - Minimalist design
	if m.engine.IsFinished {
		wpm, accuracy, _ := m.engine.GetStats()

		b.WriteString("\n\n")
		b.WriteString(m.theme.Stats.Render(fmt.Sprintf("%.0f WPM", wpm)))
		b.WriteString("  ")
		b.WriteString(m.theme.Dim.Render(fmt.Sprintf("%.1f%%", accuracy)))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Dim.Render("Press Enter to continue"))
		return b.String()
	}

	// Only show the text, no header, no stats
	b.WriteString("\n\n")

	// Render text
	targetText := m.engine.TargetText
	userInput := m.engine.UserInput

	for i := 0; i < len(targetText); i++ {
		if i < len(userInput) {
			if userInput[i] == targetText[i] {
				b.WriteString(m.theme.Correct.Render(string(targetText[i])))
			} else {
				b.WriteString(m.theme.Incorrect.Render(string(userInput[i])))
			}
		} else if i == len(userInput) {
			b.WriteString(m.theme.Cursor.Render(string(targetText[i])))
		} else {
			b.WriteString(m.theme.Dim.Render(string(targetText[i])))
		}
	}

	if len(userInput) >= len(targetText) {
		b.WriteString(m.theme.Cursor.Render(" "))
	}

	b.WriteString("\n\n")

	return b.String()
}
