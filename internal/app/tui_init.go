package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"kata/pkg/config"
	"kata/pkg/engine"
	"kata/pkg/generator"
	"kata/pkg/stats"
	"kata/pkg/themes"
)

// New creates a new TUI application model starting at the main menu
func New() tea.Model {
	return initialModel()
}

// NewPractice creates a new TUI application model starting directly in practice mode
func NewPractice(targetText string) tea.Model {
	m := initialModel()
	m.targetText = strings.TrimSpace(targetText)
	m.startPractice()
	return m
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
		// Log but don't crash, stats will just be disabled
		// fmt.Printf("Warning: Could not open database at %s: %v\n", cfg.DBPath, err)
	}

	// Load theme from config
	selectedTheme := themes.GetTheme(cfg.Theme)

	// Set generator language
	gen.SetLanguage(generator.Language(cfg.Language))

	ti := textinput.New()
	ti.Placeholder = "/path/to/file.txt"
	ti.CharLimit = 156
	ti.Width = 40

	return model{
		screen:      screenMenu,
		menuIndex:   0,
		menuOptions: []string{"Bigrams", "Keywords", "Symbols", "Code Snippets", "Timed Test", "Practice Weaknesses", "Load File", "View Stats", "Change Theme", "Change Language", "Toggle Zen Mode", "Quit"},
		generator:   gen,
		db:          db,
		textInput:   ti,
		theme:       selectedTheme,
		themeIndex:  0,
		config:      cfg,
	}
}

func (m *model) startGeneratedLesson(lt generator.LessonType, length int) {
	m.timeLimit = 0
	m.regen = func() string {
		return strings.TrimSpace(m.generator.GenerateLesson(lt, length))
	}
	m.targetText = m.regen()
	m.startPractice()
}

func (m *model) startTimedTest(limit time.Duration) {
	m.timeLimit = limit
	m.regen = func() string {
		return strings.TrimSpace(m.generator.GenerateLesson(generator.TypeWords, 220))
	}
	m.targetText = m.regen()
	m.startPractice()
}

func (m *model) startWordTest(words int) {
	m.timeLimit = 0
	m.regen = func() string {
		return strings.TrimSpace(m.generator.GenerateLesson(generator.TypeWords, words))
	}
	m.targetText = m.regen()
	m.startPractice()
}

func (m *model) weaknessText() string {
	if m.db == nil {
		return strings.TrimSpace(m.generator.GenerateLesson(generator.TypeWords, 15))
	}

	dueKeys, err := m.db.GetDueKeys(10)
	if err != nil || len(dueKeys) == 0 {
		weakKeys, err := m.db.GetWeakestKeys(10)
		if err != nil || len(weakKeys) == 0 {
			return strings.TrimSpace(m.generator.GenerateLesson(generator.TypeWords, 15))
		}
		dueKeys = weakKeys
	}

	var weakList []generator.WeakKey
	for _, k := range dueKeys {
		total := float64(k.Errors + k.Successes)
		if total == 0 {
			continue
		}
		weakList = append(weakList, generator.WeakKey{
			Key:       k.Key,
			ErrorRate: float64(k.Errors) / total,
		})
	}

	return strings.TrimSpace(m.generator.GenerateWeaknessLesson(weakList, 20))
}

func (m *model) generateWeaknessLesson() {
	m.timeLimit = 0
	m.regen = m.weaknessText
	m.targetText = m.weaknessText()
	m.startPractice()
}

func (m *model) startPractice() {
	m.screen = screenPractice
	m.engine = engine.New(m.targetText)
}

func (m *model) saveSession() {
	if m.db == nil {
		return
	}

	wpm, accuracy, duration := m.engine.GetStats()

	session := stats.Session{
		Text:       string(m.engine.TargetText),
		WPM:        wpm,
		Accuracy:   accuracy,
		Duration:   duration,
		ErrorCount: m.engine.ErrorCount,
		Timestamp:  time.Now(),
	}

	m.db.SaveSession(session)

	// Update key statistics for SRS
	m.db.UpdateKeyStats(string(m.engine.TargetText), string(m.engine.UserInput))
}
