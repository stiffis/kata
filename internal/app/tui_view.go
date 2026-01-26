package app

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
	case screenLanguageSelect:
		return m.renderLanguageSelect()
	case screenLoadFile:
		return m.renderLoadFile()
	}
	return ""
}
