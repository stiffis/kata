package keyboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"kata/pkg/stats"
)

// QWERTY keyboard layout
var keyboardLayout = [][]string{
	{"`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "="},
	{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "[", "]", "\\"},
	{"a", "s", "d", "f", "g", "h", "j", "k", "l", ";", "'"},
	{"z", "x", "c", "v", "b", "n", "m", ",", ".", "/"},
}

// GetErrorRate returns error rate for a key (0.0 to 1.0)
func GetErrorRate(key string, keyStats []stats.KeyStat) float64 {
	for _, stat := range keyStats {
		if strings.ToLower(stat.Key) == strings.ToLower(key) {
			total := stat.Errors + stat.Successes
			if total == 0 {
				return 0
			}
			return float64(stat.Errors) / float64(total)
		}
	}
	return 0
}

// GetColorForRate returns a lipgloss color based on error rate
func GetColorForRate(rate float64) lipgloss.Color {
	if rate == 0 {
		return lipgloss.Color("240") // Gray - no data
	} else if rate < 0.05 {
		return lipgloss.Color("#a6e3a1") // Green - excellent
	} else if rate < 0.15 {
		return lipgloss.Color("#94e2d5") // Teal - good
	} else if rate < 0.25 {
		return lipgloss.Color("#f9e2af") // Yellow - okay
	} else if rate < 0.40 {
		return lipgloss.Color("#fab387") // Orange - needs work
	} else {
		return lipgloss.Color("#f38ba8") // Red - problematic
	}
}

// RenderHeatmap creates an ASCII heatmap of the keyboard
func RenderHeatmap(keyStats []stats.KeyStat, theme lipgloss.Style) string {
	var b strings.Builder

	b.WriteString("\n")
	
	// Render each row
	for rowIdx, row := range keyboardLayout {
		// Add indentation for proper keyboard shape
		indent := ""
		if rowIdx == 1 {
			indent = "  "
		} else if rowIdx == 2 {
			indent = "    "
		} else if rowIdx == 3 {
			indent = "      "
		}
		
		b.WriteString(indent)
		
		for _, key := range row {
			rate := GetErrorRate(key, keyStats)
			color := GetColorForRate(rate)
			style := lipgloss.NewStyle().Foreground(color)
			
			// Format key display
			keyDisplay := fmt.Sprintf(" %s ", key)
			b.WriteString(style.Render(keyDisplay))
		}
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	
	// Legend
	b.WriteString("  Legend: ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Render("●") + " <5%  ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#94e2d5")).Render("●") + " <15%  ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render("●") + " <25%  ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")).Render("●") + " <40%  ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render("●") + " 40%+  ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("●") + " no data")
	b.WriteString("\n")

	return b.String()
}

// RenderCompactHeatmap creates a more compact version
func RenderCompactHeatmap(keyStats []stats.KeyStat) string {
	var b strings.Builder

	for rowIdx, row := range keyboardLayout {
		indent := strings.Repeat(" ", rowIdx)
		b.WriteString(indent)
		
		for _, key := range row {
			rate := GetErrorRate(key, keyStats)
			color := GetColorForRate(rate)
			style := lipgloss.NewStyle().Foreground(color).Bold(true)
			
			b.WriteString(style.Render("█"))
		}
		b.WriteString("\n")
	}

	return b.String()
}
