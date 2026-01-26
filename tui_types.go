package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"kata/pkg/config"
	"kata/pkg/engine"
	"kata/pkg/generator"
	"kata/pkg/stats"
	"kata/pkg/themes"
)

var (
	// Deprecated: These will be replaced by theme system
	// Kept for backward compatibility if any legacy code remains
	menuStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
)

type screen int

const (
	screenMenu screen = iota
	screenPractice
	screenStats
	screenThemeSelect
	screenLanguageSelect
	screenLoadFile
)

type model struct {
	screen      screen
	menuIndex   int
	menuOptions []string

	// Engine handles the typing state
	engine     *engine.Engine
	targetText string // Temporary holder for text before engine start

	// File loading
	textInput textinput.Model
	errMsg    string

	// Window dimensions
	width  int
	height int

	generator  *generator.Generator
	db         *stats.DB
	theme      themes.Theme
	themeIndex int
	config     config.Config
}
