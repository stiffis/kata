package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kata/pkg/config"
	"kata/pkg/export"
	"kata/pkg/generator"
	"kata/pkg/stats"
)

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
	// Load config to get DB path
	cfg, _ := config.Load()

	db, err := stats.NewDB(cfg.DBPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("📊 KATA Statistics")
	fmt.Println()

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

	// Load config for DB path in case we need weaknesses
	cfg, _ := config.Load()

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
		db, err := stats.NewDB(cfg.DBPath)
		if err != nil {
			fmt.Printf("Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		dueKeys, err := db.GetDueKeys(10)
		if err != nil || len(dueKeys) == 0 {
			weakKeys, err := db.GetWeakestKeys(10)
			if err != nil || len(weakKeys) == 0 {
				targetText = strings.TrimSpace(gen.GenerateLesson(generator.TypeWords, 15))
			} else {
				dueKeys = weakKeys
			}
		}

		if len(dueKeys) > 0 {
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
	// Load config to get DB path
	cfg, _ := config.Load()

	db, err := stats.NewDB(cfg.DBPath)
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
