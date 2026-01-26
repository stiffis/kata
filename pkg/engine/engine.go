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

	oldLength := len(e.UserInput)

	switch msg.String() {
	case "backspace":
		if len(e.UserInput) > 0 {
			e.UserInput = e.UserInput[:len(e.UserInput)-1]
		}
	case "ctrl+backspace", "ctrl+h", "ctrl+w":
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

	e.updateErrorsIncremental(oldLength)
	e.checkCompletion()
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
	minLength := min(len(e.UserInput), len(e.TargetText))

	for i := 0; i < minLength; i++ {
		if e.UserInput[i] != e.TargetText[i] {
			e.ErrorCount++
		}
	}

	if len(e.UserInput) > len(e.TargetText) {
		e.ErrorCount += len(e.UserInput) - len(e.TargetText)
	}
}

func (e *Engine) updateErrorsIncremental(oldLength int) {
	newLength := len(e.UserInput)

	if newLength < oldLength {
		e.calculateErrors()
		return
	}

	if newLength == oldLength {
		return
	}

	for i := oldLength; i < newLength; i++ {
		if i < len(e.TargetText) {
			if e.UserInput[i] != e.TargetText[i] {
				e.ErrorCount++
			}
		} else {
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

	// Calculate WPM based on correct characters typed, not total target length
	// Standard: 1 word = 5 characters
	correctChars := max(0, len(e.TargetText)-e.ErrorCount)
	words := float64(correctChars) / 5.0
	wpm = (words / duration) * 60.0

	// Calculate accuracy based on actual attempts made
	totalAttempts := len(e.UserInput)
	if totalAttempts == 0 {
		accuracy = 100.0
	} else {
		correctAttempts := max(0, totalAttempts-e.ErrorCount)
		accuracy = float64(correctAttempts) / float64(totalAttempts) * 100.0
	}

	return wpm, accuracy, duration
}
