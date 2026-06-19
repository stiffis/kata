package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"kata/pkg/config"
	"kata/pkg/generator"
	"kata/pkg/themes"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.screen == screenPractice && m.timeLimit > 0 && m.engine != nil &&
			!m.engine.IsFinished && !m.engine.StartTime.IsZero() {
			if time.Since(m.engine.StartTime) >= m.timeLimit {
				m.engine.Finish()
				m.saveSession()
			}
		}
		return m, tickCmd()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.screen == screenStats && !m.statsReady {
			m.statsViewport = viewport.New(msg.Width, msg.Height-2)
			m.statsViewport.YPosition = 0
			m.statsViewport.SetContent(m.buildStatsContent())
			m.statsReady = true
		} else if m.screen == screenStats && m.statsReady {
			m.statsViewport.Width = msg.Width
			m.statsViewport.Height = msg.Height - 2
		}

		return m, nil
	case tea.KeyMsg:
		if m.screen == screenStats && m.statsReady {
			var cmd tea.Cmd
			m.statsViewport, cmd = m.statsViewport.Update(msg)

			switch msg.String() {
			case "ctrl+c", "q":
				if m.db != nil {
					m.db.Close()
				}
				return m, tea.Quit
			case "esc", "enter":
				m.screen = screenMenu
				m.statsReady = false
				return m, nil
			}
			return m, cmd
		}

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
		case screenModeSelect:
			return m.handleModeSelectInput(msg)
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
	case 0:
		m.startGeneratedLesson(generator.TypeBigrams, 20)
	case 1:
		m.startGeneratedLesson(generator.TypeWords, 15)
	case 2:
		m.startGeneratedLesson(generator.TypeSymbols, 10)
	case 3:
		m.startGeneratedLesson(generator.TypeCode, 2)
	case 4:
		m.screen = screenModeSelect
		m.modeIndex = 0
		return m, nil
	case 5:
		m.generateWeaknessLesson()
	case 6:
		m.screen = screenLoadFile
		m.textInput.Focus()
		m.textInput.SetValue("")
		m.errMsg = ""
		return m, textinput.Blink
	case 7:
		m.screen = screenStats
		m.statsReady = false
		if m.width > 0 && m.height > 0 {
			m.statsViewport = viewport.New(m.width, m.height-2)
			m.statsViewport.YPosition = 0
			m.statsViewport.SetContent(m.buildStatsContent())
			m.statsReady = true
		}
		return m, nil
	case 8:
		m.screen = screenThemeSelect
		m.themeIndex = 0
		return m, nil
	case 9:
		m.screen = screenLanguageSelect
		m.themeIndex = 0
		return m, nil
	case 10:
		m.config.ZenMode = !m.config.ZenMode
		if err := config.Save(m.config); err != nil {
			fmt.Printf("Warning: Could not save config: %v\n", err)
		}
		return m, nil
	case 11:
		if m.db != nil {
			m.db.Close()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handlePracticeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.engine.IsFinished {
		switch msg.String() {
		case "q", "ctrl+c":
			if m.db != nil {
				m.db.Close()
			}
			return m, tea.Quit
		case "enter":
			m.screen = screenMenu
			m.menuIndex = 0
			return m, nil
		case "r", "tab":
			m.startPractice()
			return m, nil
		case "n":
			if m.regen != nil {
				m.targetText = m.regen()
			}
			m.startPractice()
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

type practiceMode struct {
	label   string
	timed   bool
	seconds int
	words   int
}

func practiceModes() []practiceMode {
	return []practiceMode{
		{label: "15 seconds", timed: true, seconds: 15},
		{label: "30 seconds", timed: true, seconds: 30},
		{label: "60 seconds", timed: true, seconds: 60},
		{label: "10 words", words: 10},
		{label: "25 words", words: 25},
		{label: "50 words", words: 50},
	}
}

func (m *model) applyMode(i int) {
	modes := practiceModes()
	if i < 0 || i >= len(modes) {
		return
	}
	mode := modes[i]
	if mode.timed {
		m.startTimedTest(time.Duration(mode.seconds) * time.Second)
	} else {
		m.startWordTest(mode.words)
	}
}

func (m model) handleModeSelectInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIndex := len(practiceModes()) - 1

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
		if m.modeIndex > 0 {
			m.modeIndex--
		} else {
			m.modeIndex = maxIndex
		}
	case "down", "j":
		if m.modeIndex < maxIndex {
			m.modeIndex++
		} else {
			m.modeIndex = 0
		}
	case "enter":
		m.applyMode(m.modeIndex)
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
