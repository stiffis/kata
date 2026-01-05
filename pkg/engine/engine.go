package engine

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Engine struct {
	TargetText []rune
	UserInput  []rune
	StartTime  time.Time
	EndTime    time.Time
	IsFinished bool
	ErrorCount int
}

func New(targetText string) *Engine {
	return &Engine{
		TargetText: []rune(targetText),
		UserInput:  []rune{},
		IsFinished: false,
	}
}

func (e *Engine) ProcessKey(msg tea.KeyMsg) {
	if e.IsFinished {
		return
	}

	if e.StartTime.IsZero() {
		e.StartTime = time.Now()
	}

	switch msg.String() {
	case "backspace":
		if len(e.UserInput) > 0 {
			e.UserInput = e.UserInput[:len(e.UserInput)-1]
		}
	case "ctrl+backspace", "ctrl+h":
		e.UserInput = e.deleteLastWord(e.UserInput)
	case "enter":
		e.UserInput = append(e.UserInput, '\n')
	case "tab":
		e.UserInput = append(e.UserInput, '\t')
	default:
		if len(msg.Runes) > 0 {
			e.UserInput = append(e.UserInput, msg.Runes...)
		} else if len(msg.String()) == 1 {
			e.UserInput = append(e.UserInput, rune(msg.String()[0]))
		}
	}

	e.checkCompletion()
	e.calculateErrors()
}

func (e *Engine) checkCompletion() {
	if len(e.UserInput) >= len(e.TargetText) {
		match := true
		for i := 0; i < len(e.TargetText); i++ {
			if e.UserInput[i] != e.TargetText[i] {
				match = false
				break
			}
		}

		if match {
			e.IsFinished = true
			e.EndTime = time.Now()
		}
	}
}

func (e *Engine) calculateErrors() {
	e.ErrorCount = 0
	checkLength := len(e.UserInput)
	if checkLength > len(e.TargetText) {
		checkLength = len(e.TargetText)
	}
	for i := 0; i < checkLength; i++ {
		if e.UserInput[i] != e.TargetText[i] {
			e.ErrorCount++
		}
	}
}

func (e *Engine) deleteLastWord(input []rune) []rune {
	if len(input) == 0 {
		return input
	}

	endIdx := len(input) - 1
	for endIdx >= 0 {
		r := input[endIdx]
		if r != ' ' && r != '\t' && r != '\n' {
			break
		}
		endIdx--
	}

	if endIdx < 0 {
		return []rune{}
	}

	startIdx := endIdx
	for startIdx >= 0 {
		r := input[startIdx]
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		startIdx--
	}

	return input[:startIdx+1]
}

func (e *Engine) GetStats() (wpm float64, accuracy float64, duration float64) {
	if e.StartTime.IsZero() {
		return 0, 0, 0
	}

	endTime := e.EndTime
	if endTime.IsZero() {
		endTime = time.Now()
	}

	duration = endTime.Sub(e.StartTime).Seconds()

	if duration == 0 {
		duration = 1 // minimal duration
	}

	words := float64(len(e.TargetText)) / 5.0
	wpm = (words / duration) * 60.0

	if len(e.TargetText) > 0 {
		accuracy = float64(len(e.TargetText)-e.ErrorCount) / float64(len(e.TargetText)) * 100.0
	}

	return wpm, accuracy, duration
}

