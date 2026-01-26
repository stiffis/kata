package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/guptarohit/asciigraph"
	"golang.org/x/term"

	"kata/pkg/keyboard"
)

func (m model) renderStats() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("📊 Your Statistics"))
	b.WriteString("\n\n")

	if m.db == nil {
		b.WriteString(m.theme.Incorrect.Render("No database connection available"))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Dim.Render("Press ESC to return to menu"))
		return b.String()
	}

	// Get terminal width
	termWidth := 80 // default
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		termWidth = width
	}
	separator := m.theme.Separator.Render(strings.Repeat("─", termWidth))

	// Get sessions for graphing
	sessions, err := m.db.GetSessionsForGraph(20)
	if err != nil || len(sessions) == 0 {
		b.WriteString(m.theme.Dim.Render("No session data yet. Complete some practice sessions!"))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Dim.Render("Press ESC or Enter to return to menu"))
		return b.String()
	}

	// WPM Over Time Graph
	b.WriteString(m.theme.Stats.Render("📈 WPM Progress Over Time:"))
	b.WriteString("\n")

	wpmData := make([]float64, len(sessions))
	for i, s := range sessions {
		wpmData[i] = s.WPM
	}

	graph := asciigraph.Plot(wpmData,
		asciigraph.Height(8),
		asciigraph.Width(60),
		asciigraph.Caption(fmt.Sprintf("Last %d sessions", len(sessions))))

	b.WriteString(m.theme.Correct.Render(graph))
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n\n")

	// Average WPM
	avgWPM, _ := m.db.GetAverageWPM()
	if avgWPM > 0 {
		b.WriteString(m.theme.Stats.Render(fmt.Sprintf("Average WPM: %.0f", avgWPM)))
		b.WriteString("\n")
		b.WriteString(separator)
		b.WriteString("\n\n")
	}

	// Accuracy sparkline
	b.WriteString(m.theme.Stats.Render("🎯 Accuracy Trend:"))
	b.WriteString("\n")

	accData := make([]float64, len(sessions))
	for i, s := range sessions {
		accData[i] = s.Accuracy
	}

	accGraph := asciigraph.Plot(accData,
		asciigraph.Height(6),
		asciigraph.Width(60),
		asciigraph.Caption("Accuracy %"))

	b.WriteString(menuStyle.Render(accGraph))
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n\n")

	// Weakest Keys Bar Chart
	weakKeys, err := m.db.GetWeakestKeys(8)
	if err == nil && len(weakKeys) > 0 {
		b.WriteString(m.theme.Incorrect.Render("🔥 Your Weakest Keys:"))
		b.WriteString("\n")

		maxErrors := 1 // Prevent division by zero
		for _, k := range weakKeys {
			if k.Errors > maxErrors {
				maxErrors = k.Errors
			}
		}

		for i, k := range weakKeys {
			if i >= 5 {
				break
			}
			total := k.Errors + k.Successes
			if total == 0 {
				continue // Skip if no data
			}

			errorRate := float64(k.Errors) / float64(total) * 100.0

			keyDisplay := k.Key
			if k.Key == "\n" {
				keyDisplay = "↵"
			} else if k.Key == "\t" {
				keyDisplay = "⭾"
			} else if k.Key == " " {
				keyDisplay = "␣"
			}

			// Create horizontal bar (ensure it's never negative)
			barLength := int(float64(k.Errors) / float64(maxErrors) * 30)
			if barLength < 0 {
				barLength = 0
			}
			if barLength < 1 && k.Errors > 0 {
				barLength = 1
			}
			bar := strings.Repeat("█", barLength)

			b.WriteString(fmt.Sprintf("  '%s' %s %.0f%% (%d errors)\n",
				keyDisplay,
				m.theme.Incorrect.Render(bar),
				errorRate,
				k.Errors))
		}
		b.WriteString("\n")
		b.WriteString(m.theme.Correct.Render("💡 Tip: Use 'Practice Weaknesses' to improve!"))
		b.WriteString("\n")
		b.WriteString(separator)
		b.WriteString("\n")
	}

	dueKeys, err := m.db.GetDueKeys(100)
	if err == nil {
		b.WriteString("\n")
		dueCount := len(dueKeys)
		if dueCount > 0 {
			b.WriteString(m.theme.Stats.Render(fmt.Sprintf("🔄 Keys Due for Review: %d", dueCount)))
		} else {
			b.WriteString(m.theme.Dim.Render("🔄 No keys due for review - great job! 🎉"))
		}
		b.WriteString("\n")
		b.WriteString(separator)
		b.WriteString("\n")
	}

	// Recent sessions mini summary
	recentSessions, _ := m.db.GetRecentSessions(3)
	if len(recentSessions) > 0 {
		b.WriteString("\n")
		b.WriteString(m.theme.Dim.Render("Recent Sessions:"))
		b.WriteString("\n")
		for _, s := range recentSessions {
			timeStr := s.Timestamp.Format("Jan 02 15:04")
			wpmIndicator := "→"
			if avgWPM > 0 {
				if s.WPM >= avgWPM {
					wpmIndicator = m.theme.Correct.Render("↑")
				} else {
					wpmIndicator = m.theme.Incorrect.Render("↓")
				}
			}
			b.WriteString(fmt.Sprintf("  %s %s WPM: %.0f | Acc: %.1f%%\n",
				m.theme.Dim.Render(timeStr), wpmIndicator, s.WPM, s.Accuracy))
		}
	}

	// Keyboard Heatmap
	allKeyStats, err := m.db.GetAllKeyStats()
	if err == nil && len(allKeyStats) > 0 {
		b.WriteString("\n")
		b.WriteString(separator)
		b.WriteString("\n")
		b.WriteString(m.theme.Menu.Render("⌨️  Keyboard Heatmap:"))
		b.WriteString("\n")
		heatmap := keyboard.RenderHeatmap(allKeyStats, m.theme.Dim)
		b.WriteString(heatmap)
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("Press ESC or Enter to return to menu"))

	return b.String()
}
