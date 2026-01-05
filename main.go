package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"kata/pkg/config"
	"kata/pkg/export"
	"kata/pkg/generator"
	"kata/pkg/keyboard"
	"kata/pkg/stats"
	"kata/pkg/themes"
)

var (
	// Deprecated: These will be replaced by theme system
	correctStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	incorrectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Underline(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	statsStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	menuStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type screen int

const (
	screenMenu screen = iota
	screenPractice
	screenStats
	screenThemeSelect
)

type model struct {
	screen      screen
	menuIndex   int
	menuOptions []string
	
	targetText  string
	userInput   string
	startTime   time.Time
	endTime     time.Time
	finished    bool
	errorCount  int
	
	generator   *generator.Generator
	db          *stats.DB
	theme       themes.Theme
	themeIndex  int
	config      config.Config
}

func initialModel() model {
	gen := generator.New()
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}
	
	db, err := stats.NewDB(cfg.DBPath)
	if err != nil {
		fmt.Printf("Warning: Could not open database: %v\n", err)
	}
	
	// Load theme from config
	selectedTheme := themes.GetTheme(cfg.Theme)
	
	return model{
		screen:      screenMenu,
		menuIndex:   0,
		menuOptions: []string{"Bigrams", "Keywords", "Symbols", "Code Snippets", "Practice Weaknesses", "Load File", "View Stats", "Change Theme", "Toggle Zen Mode", "Quit"},
		generator:   gen,
		db:          db,
		theme:       selectedTheme,
		themeIndex:  0,
		config:      cfg,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case screenMenu:
			return m.handleMenuInput(msg)
		case screenPractice:
			return m.handlePracticeInput(msg)
		case screenStats:
			return m.handleStatsInput(msg)
		case screenThemeSelect:
			return m.handleThemeSelectInput(msg)
		}
	}
	return m, nil
}

func (m model) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.db != nil {
			m.db.Close()
		}
		return m, tea.Quit
	case "up", "k":
		if m.menuIndex > 0 {
			m.menuIndex--
		}
	case "down", "j":
		if m.menuIndex < len(m.menuOptions)-1 {
			m.menuIndex++
		}
	case "enter":
		return m.selectMenuItem()
	}
	return m, nil
}

func (m model) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.menuIndex {
	case 0: // Bigrams
		m.targetText = strings.TrimSpace(m.generator.GenerateLesson(generator.TypeBigrams, 20))
		m.startPractice()
	case 1: // Keywords
		m.targetText = strings.TrimSpace(m.generator.GenerateLesson(generator.TypeWords, 15))
		m.startPractice()
	case 2: // Symbols
		m.targetText = strings.TrimSpace(m.generator.GenerateLesson(generator.TypeSymbols, 10))
		m.startPractice()
	case 3: // Code Snippets
		m.targetText = strings.TrimSpace(m.generator.GenerateLesson(generator.TypeCode, 2))
		m.startPractice()
	case 4: // Practice Weaknesses
		m.generateWeaknessLesson()
	case 5: // Load File
		if len(os.Args) > 1 {
			content, err := m.generator.GenerateFromFile(os.Args[1])
			if err == nil {
				m.targetText = strings.TrimSpace(content)
				m.startPractice()
			}
		}
	case 6: // View Stats
		m.screen = screenStats
		return m, nil
	case 7: // Change Theme
		m.screen = screenThemeSelect
		m.themeIndex = 0
		return m, nil
	case 8: // Toggle Zen Mode
		m.config.ZenMode = !m.config.ZenMode
		if err := config.Save(m.config); err != nil {
			fmt.Printf("Warning: Could not save config: %v\n", err)
		}
		return m, nil
	case 9: // Quit
		if m.db != nil {
			m.db.Close()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) generateWeaknessLesson() {
	if m.db == nil {
		// Fallback to keywords
		m.targetText = m.generator.GenerateLesson(generator.TypeWords, 15)
		m.targetText = strings.TrimSpace(m.targetText)
		m.startPractice()
		return
	}

	weakKeys, err := m.db.GetWeakestKeys(10)
	if err != nil || len(weakKeys) == 0 {
		// Fallback to keywords if no data yet
		m.targetText = m.generator.GenerateLesson(generator.TypeWords, 15)
		m.targetText = strings.TrimSpace(m.targetText)
		m.startPractice()
		return
	}

	// Convert to generator format
	var weakList []generator.WeakKey
	for _, k := range weakKeys {
		total := float64(k.Errors + k.Successes)
		if total == 0 {
			continue
		}
		errorRate := float64(k.Errors) / total
		weakList = append(weakList, generator.WeakKey{
			Key:       k.Key,
			ErrorRate: errorRate,
		})
	}

	m.targetText = m.generator.GenerateWeaknessLesson(weakList, 20)
	m.targetText = strings.TrimSpace(m.targetText)
	m.startPractice()
}

func (m *model) startPractice() {
	m.screen = screenPractice
	m.userInput = ""
	m.startTime = time.Time{}
	m.finished = false
	m.errorCount = 0
}

func (m model) handlePracticeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.finished {
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			if m.db != nil {
				m.db.Close()
			}
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			m.screen = screenMenu
			m.menuIndex = 0
			return m, nil
		}
		return m, nil
	}

	// Start timer on first keystroke
	if m.startTime.IsZero() {
		m.startTime = time.Now()
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = screenMenu
		return m, nil
	case "ctrl+z":
		// Toggle Zen mode during practice
		m.config.ZenMode = !m.config.ZenMode
		return m, nil
	case "backspace":
		if len(m.userInput) > 0 {
			m.userInput = m.userInput[:len(m.userInput)-1]
		}
	case "ctrl+backspace", "ctrl+h":
		// Delete entire word
		m.userInput = m.deleteLastWord(m.userInput)
	case "enter":
		m.userInput += "\n"
	case "tab":
		m.userInput += "\t"
	default:
		if len(msg.String()) == 1 {
			m.userInput += msg.String()
		}
	}

	// Check if finished
	if len(m.userInput) >= len(m.targetText) {
		// User typed at least as much as target
		if m.userInput[:len(m.targetText)] == m.targetText {
			m.finished = true
			m.endTime = time.Now()
			m.saveSession()
		}
	}

	// Count errors (only within target text length)
	m.errorCount = 0
	checkLength := len(m.userInput)
	if checkLength > len(m.targetText) {
		checkLength = len(m.targetText)
	}
	for i := 0; i < checkLength; i++ {
		if m.userInput[i] != m.targetText[i] {
			m.errorCount++
		}
	}

	return m, nil
}

func (m *model) saveSession() {
	if m.db == nil {
		return
	}

	duration := m.endTime.Sub(m.startTime).Seconds()
	words := float64(len(m.targetText)) / 5.0
	wpm := (words / duration) * 60.0
	accuracy := float64(len(m.targetText)-m.errorCount) / float64(len(m.targetText)) * 100.0

	session := stats.Session{
		Text:       m.targetText,
		WPM:        wpm,
		Accuracy:   accuracy,
		Duration:   duration,
		ErrorCount: m.errorCount,
		Timestamp:  time.Now(),
	}

	m.db.SaveSession(session)
	
	// Update key statistics for SRS
	m.db.UpdateKeyStats(m.targetText, m.userInput)
}

func (m *model) deleteLastWord(input string) string {
	if len(input) == 0 {
		return input
	}

	// Trim trailing spaces first
	trimmed := strings.TrimRight(input, " \t\n")
	if len(trimmed) == 0 {
		return ""
	}

	// Find the last space/tab/newline before the word
	lastSpace := -1
	for i := len(trimmed) - 1; i >= 0; i-- {
		if trimmed[i] == ' ' || trimmed[i] == '\t' || trimmed[i] == '\n' {
			lastSpace = i
			break
		}
	}

	// If no space found, delete everything
	if lastSpace == -1 {
		return ""
	}

	// Keep text up to and including the last space
	return trimmed[:lastSpace+1]
}

func (m model) handleThemeSelectInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	themeNames := themes.ListThemes()
	maxIndex := len(themeNames) - 1
	
	switch msg.String() {
	case "ctrl+c", "q":
		if m.db != nil {
			m.db.Close()
		}
		return m, tea.Quit
	case "esc":
		m.screen = screenMenu
		return m, nil
	case "up", "k":
		if m.themeIndex > 0 {
			m.themeIndex--
		} else {
			m.themeIndex = maxIndex // Wrap to bottom
		}
	case "down", "j":
		if m.themeIndex < maxIndex {
			m.themeIndex++
		} else {
			m.themeIndex = 0 // Wrap to top
		}
	case "enter":
		// Apply selected theme
		if m.themeIndex >= 0 && m.themeIndex < len(themeNames) {
			themeName := themeNames[m.themeIndex]
			m.theme = themes.GetTheme(themeName)
			
			// Save theme to config
			m.config.Theme = themeName
			if err := config.Save(m.config); err != nil {
				// Continue even if save fails
				fmt.Printf("Warning: Could not save config: %v\n", err)
			}
		}
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func (m model) handleStatsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.db != nil {
			m.db.Close()
		}
		return m, tea.Quit
	case "esc", "enter":
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenMenu:
		return m.renderMenu()
	case screenPractice:
		return m.renderPractice()
	case screenStats:
		return m.renderStats()
	case screenThemeSelect:
		return m.renderThemeSelect()
	}
	return ""
}

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

	return b.String()
}

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

func (m model) renderPractice() string {
	// Use Zen mode if enabled
	if m.config.ZenMode {
		return m.renderPracticeZen()
	}
	
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("🥋 KATA - The Way of the Keyboard"))
	b.WriteString("\n\n")

	if m.finished {
		duration := m.endTime.Sub(m.startTime).Seconds()
		words := float64(len(m.targetText)) / 5.0
		wpm := (words / duration) * 60.0
		accuracy := float64(len(m.targetText)-m.errorCount) / float64(len(m.targetText)) * 100.0

		b.WriteString(m.theme.Stats.Render(fmt.Sprintf("✓ Complete!\n\n")))
		b.WriteString(m.theme.Stats.Render(fmt.Sprintf("Time: %.1f seconds\n", duration)))
		b.WriteString(m.theme.Stats.Render(fmt.Sprintf("WPM: %.0f\n", wpm)))
		b.WriteString(m.theme.Stats.Render(fmt.Sprintf("Accuracy: %.1f%%\n", accuracy)))
		b.WriteString("\n")
		b.WriteString(m.theme.Dim.Render("Press Enter to return to menu | q to quit"))
		return b.String()
	}

	// Render the text with colors
	for i := 0; i < len(m.targetText); i++ {
		if i < len(m.userInput) {
			if m.userInput[i] == m.targetText[i] {
				b.WriteString(m.theme.Correct.Render(string(m.targetText[i])))
			} else {
				b.WriteString(m.theme.Incorrect.Render(string(m.targetText[i])))
			}
		} else if i == len(m.userInput) {
			b.WriteString(m.theme.Cursor.Render(string(m.targetText[i])))
		} else {
			b.WriteString(m.theme.Dim.Render(string(m.targetText[i])))
		}
	}
	
	// Show cursor if user typed past the end
	if len(m.userInput) >= len(m.targetText) {
		b.WriteString(m.theme.Cursor.Render(" "))
	}

	b.WriteString("\n\n")
	
	if !m.startTime.IsZero() {
		elapsed := time.Since(m.startTime).Seconds()
		progress := float64(len(m.userInput)) / float64(len(m.targetText)) * 100.0
		if progress > 100.0 {
			progress = 100.0
		}
		b.WriteString(m.theme.Dim.Render(fmt.Sprintf("Progress: %.0f%% | Time: %.0fs | Errors: %d", progress, elapsed, m.errorCount)))
	} else {
		b.WriteString(m.theme.Dim.Render("Start typing to begin..."))
	}

	b.WriteString("\n\n")
	b.WriteString(m.theme.Dim.Render("ESC to menu | Ctrl+Z to toggle zen | Ctrl+C to quit"))

	return b.String()
}

func (m model) renderPracticeZen() string {
	var b strings.Builder

	// Zen Mode - Minimalist design
	if m.finished {
		duration := m.endTime.Sub(m.startTime).Seconds()
		words := float64(len(m.targetText)) / 5.0
		wpm := (words / duration) * 60.0
		accuracy := float64(len(m.targetText)-m.errorCount) / float64(len(m.targetText)) * 100.0

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
	for i := 0; i < len(m.targetText); i++ {
		if i < len(m.userInput) {
			if m.userInput[i] == m.targetText[i] {
				b.WriteString(m.theme.Correct.Render(string(m.targetText[i])))
			} else {
				b.WriteString(m.theme.Incorrect.Render(string(m.targetText[i])))
			}
		} else if i == len(m.userInput) {
			b.WriteString(m.theme.Cursor.Render(string(m.targetText[i])))
		} else {
			b.WriteString(m.theme.Dim.Render(string(m.targetText[i])))
		}
	}
	
	if len(m.userInput) >= len(m.targetText) {
		b.WriteString(m.theme.Cursor.Render(" "))
	}

	b.WriteString("\n\n")

	return b.String()
}

func main() {
	// Parse CLI flags
	var (
		showStats    = false
		setTheme     = ""
		enableZen    = false
		practiceMode = ""
		practiceFile = ""
		showHelp     = false
		exportFormat = ""
		exportOutput = ""
	)
	
	// Simple flag parsing
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--stats", "-s":
			showStats = true
		case "--theme", "-t":
			if i+1 < len(os.Args) {
				setTheme = os.Args[i+1]
				i++
			}
		case "--zen", "-z":
			enableZen = true
		case "--file", "-f":
			if i+1 < len(os.Args) {
				practiceFile = os.Args[i+1]
				i++
			}
		case "export", "e":
			if i+1 < len(os.Args) {
				exportFormat = os.Args[i+1]
				i++
			}
			if i+1 < len(os.Args) {
				exportOutput = os.Args[i+1]
				i++
			}
		case "practice", "p":
			if i+1 < len(os.Args) {
				practiceMode = os.Args[i+1]
				i++
			}
		case "--help", "-h", "help":
			showHelp = true
		}
	}
	
	// Handle --help
	if showHelp {
		printHelp()
		return
	}
	
	// Load config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}
	
	// Handle --theme
	if setTheme != "" {
		cfg.Theme = setTheme
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error saving theme: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Theme set to: %s\n", setTheme)
		return
	}
	
	// Handle --zen
	if enableZen {
		cfg.ZenMode = true
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error saving zen mode: %v\n", err)
		}
	}
	
	// Handle export
	if exportFormat != "" {
		handleExport(exportFormat, exportOutput)
		return
	}
	
	// Handle --stats
	if showStats {
		printStats()
		return
	}
	
	// Handle practice mode
	if practiceMode != "" {
		runPracticeMode(practiceMode)
		return
	}
	
	// Handle practice from file
	if practiceFile != "" {
		runPracticeFromFile(practiceFile)
		return
	}
	
	// Normal interactive mode
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func printHelp() {
	help := `🥋 KATA - The Way of the Keyboard

USAGE:
    kata [OPTIONS] [COMMAND]

OPTIONS:
    --stats, -s              Show statistics and exit
    --theme, -t <name>       Set theme (default, catppuccin, rose-pine, dracula, nord, gruvbox)
    --zen, -z                Enable zen mode
    --help, -h               Show this help

COMMANDS:
    practice <mode>          Start practice directly
                            Modes: bigrams, keywords, symbols, code, weaknesses
    export <format> <file>   Export statistics to file
                            Formats: json, csv

OPTIONS WITH FILES:
    --file, -f <path>        Practice with custom file content

EXAMPLES:
    kata                     Start interactive mode
    kata --stats             Show your statistics
    kata --theme dracula     Set theme to dracula
    kata --zen               Start with zen mode enabled
    kata practice bigrams    Practice bigrams directly
    kata --file lesson.txt   Practice with custom lesson file
    kata export json stats.json   Export to JSON
    kata export csv stats.csv     Export to CSV

"Slow is smooth. Smooth is fast."
`
	fmt.Print(help)
}

func printStats() {
	db, err := stats.NewDB("kata.db")
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	
	fmt.Println("📊 KATA Statistics\n")
	
	// Average WPM
	avgWPM, err := db.GetAverageWPM()
	if err == nil && avgWPM > 0 {
		fmt.Printf("Average WPM: %.0f\n\n", avgWPM)
	}
	
	// Recent sessions
	sessions, err := db.GetRecentSessions(5)
	if err == nil && len(sessions) > 0 {
		fmt.Println("Recent Sessions:")
		for _, s := range sessions {
			fmt.Printf("  %s | WPM: %.0f | Accuracy: %.1f%%\n",
				s.Timestamp.Format("Jan 02 15:04"), s.WPM, s.Accuracy)
		}
		fmt.Println()
	}
	
	// Weakest keys
	weakKeys, err := db.GetWeakestKeys(5)
	if err == nil && len(weakKeys) > 0 {
		fmt.Println("Weakest Keys:")
		for _, k := range weakKeys {
			total := k.Errors + k.Successes
			errorRate := float64(k.Errors) / float64(total) * 100.0
			keyDisplay := k.Key
			if k.Key == "\n" {
				keyDisplay = "↵"
			} else if k.Key == "\t" {
				keyDisplay = "⭾"
			} else if k.Key == " " {
				keyDisplay = "␣"
			}
			fmt.Printf("  '%s' → %.0f%% errors (%d/%d)\n", keyDisplay, errorRate, k.Errors, total)
		}
	}
}

func runPracticeMode(mode string) {
	gen := generator.New()
	var targetText string
	
	switch mode {
	case "bigrams", "b":
		targetText = strings.TrimSpace(gen.GenerateLesson(generator.TypeBigrams, 20))
	case "keywords", "k":
		targetText = strings.TrimSpace(gen.GenerateLesson(generator.TypeWords, 15))
	case "symbols", "s":
		targetText = strings.TrimSpace(gen.GenerateLesson(generator.TypeSymbols, 10))
	case "code", "c":
		targetText = strings.TrimSpace(gen.GenerateLesson(generator.TypeCode, 2))
	case "weaknesses", "w":
		db, err := stats.NewDB("kata.db")
		if err != nil {
			fmt.Printf("Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		
		weakKeys, err := db.GetWeakestKeys(10)
		if err != nil || len(weakKeys) == 0 {
			targetText = strings.TrimSpace(gen.GenerateLesson(generator.TypeWords, 15))
		} else {
			var weakList []generator.WeakKey
			for _, k := range weakKeys {
				total := float64(k.Errors + k.Successes)
				if total == 0 {
					continue
				}
				errorRate := float64(k.Errors) / total
				weakList = append(weakList, generator.WeakKey{
					Key:       k.Key,
					ErrorRate: errorRate,
				})
			}
			targetText = strings.TrimSpace(gen.GenerateWeaknessLesson(weakList, 20))
		}
	default:
		fmt.Printf("Unknown practice mode: %s\n", mode)
		fmt.Println("Available modes: bigrams, keywords, symbols, code, weaknesses")
		os.Exit(1)
	}
	
	// Create model with target text and start practice
	m := initialModel()
	m.targetText = targetText
	m.startPractice()
	
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func handleExport(format, output string) {
	db, err := stats.NewDB("kata.db")
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Set default output filename
	if output == "" {
		output = fmt.Sprintf("kata-stats-%s.%s", time.Now().Format("2006-01-02"), format)
	}

	switch format {
	case "json":
		if err := export.ToJSON(db, output); err != nil {
			fmt.Printf("Error exporting to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Statistics exported to: %s\n", output)
	case "csv":
		if err := export.ToCSV(db, output); err != nil {
			fmt.Printf("Error exporting to CSV: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Statistics exported to: %s\n", output)
	default:
		fmt.Printf("Unknown export format: %s\n", format)
		fmt.Println("Available formats: json, csv")
		os.Exit(1)
	}
}

func runPracticeFromFile(filepath string) {
	gen := generator.New()
	content, err := gen.GenerateFromFile(filepath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	targetText := strings.TrimSpace(content)
	if len(targetText) == 0 {
		fmt.Println("Error: File is empty")
		os.Exit(1)
	}

	// Create model with target text and start practice
	m := initialModel()
	m.targetText = targetText
	m.startPractice()

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
