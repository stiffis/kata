package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"kata/pkg/config"
	"kata/pkg/generator"
	"kata/pkg/themes"
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
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
		case screenLanguageSelect:
			return m.handleLanguageSelectInput(msg)
		case screenLoadFile:
			return m.handleLoadFileInput(msg)
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
		m.screen = screenLoadFile
		m.textInput.Focus()
		m.textInput.SetValue("")
		m.errMsg = ""
		return m, textinput.Blink
	case 6: // View Stats
		m.screen = screenStats
		return m, nil
	case 7: // Change Theme
		m.screen = screenThemeSelect
		m.themeIndex = 0
		return m, nil
	case 8: // Change Language
		m.screen = screenLanguageSelect
		// Reuse themeIndex for language list navigation as it's just an int
		m.themeIndex = 0
		return m, nil
	case 9: // Toggle Zen Mode
		m.config.ZenMode = !m.config.ZenMode
		if err := config.Save(m.config); err != nil {
			fmt.Printf("Warning: Could not save config: %v\n", err)
		}
		return m, nil
	case 10: // Quit
		if m.db != nil {
			m.db.Close()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handlePracticeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.engine.IsFinished {
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
	default:
		// Delegate to engine
		m.engine.ProcessKey(msg)
	}

	// Check if just finished
	if m.engine.IsFinished {
		m.saveSession()
	}

	return m, nil
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

func (m model) handleLanguageSelectInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	languages := []string{"go", "cpp", "javascript", "rust", "python", "english", "spanish", "french", "german"}
	maxIndex := len(languages) - 1

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
		// Apply selected language
		if m.themeIndex >= 0 && m.themeIndex < len(languages) {
			lang := languages[m.themeIndex]

			// Update generator
			m.generator.SetLanguage(generator.Language(lang))

			// Save to config
			m.config.Language = lang
			if err := config.Save(m.config); err != nil {
				fmt.Printf("Warning: Could not save config: %v\n", err)
			}
		}
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func (m model) handleLoadFileInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenMenu
		return m, nil
	case tea.KeyEnter:
		filepath := m.textInput.Value()
		content, err := m.generator.GenerateFromFile(filepath)
		if err != nil {
			m.errMsg = fmt.Sprintf("Error: %v", err)
			return m, nil
		}

		if strings.TrimSpace(content) == "" {
			m.errMsg = "Error: File is empty"
			return m, nil
		}

		m.targetText = strings.TrimSpace(content)
		m.startPractice()
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
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
