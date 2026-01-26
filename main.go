package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"kata/internal/app"
	"kata/pkg/config"
)

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
	p := tea.NewProgram(app.New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
