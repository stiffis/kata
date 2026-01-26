package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"kata/pkg/config"
	"kata/pkg/engine"
	"kata/pkg/generator"
	"kata/pkg/stats"
	"kata/pkg/themes"
)

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
		menuOptions: []string{"Bigrams", "Keywords", "Symbols", "Code Snippets", "Practice Weaknesses", "Load File", "View Stats", "Change Theme", "Change Language", "Toggle Zen Mode", "Quit"},
		generator:   gen,
		db:          db,
		textInput:   ti,
		theme:       selectedTheme,
		themeIndex:  0,
		config:      cfg,
	}
}

func (m *model) generateWeaknessLesson() {
	if m.db == nil {
		m.targetText = m.generator.GenerateLesson(generator.TypeWords, 15)
		m.targetText = strings.TrimSpace(m.targetText)
		m.startPractice()
		return
	}

	dueKeys, err := m.db.GetDueKeys(10)
	if err != nil || len(dueKeys) == 0 {
		weakKeys, err := m.db.GetWeakestKeys(10)
		if err != nil || len(weakKeys) == 0 {
			m.targetText = m.generator.GenerateLesson(generator.TypeWords, 15)
			m.targetText = strings.TrimSpace(m.targetText)
			m.startPractice()
			return
		}
		dueKeys = weakKeys
	}

	var weakList []generator.WeakKey
	for _, k := range dueKeys {
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
