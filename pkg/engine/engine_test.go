package engine

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEngineTyping(t *testing.T) {
	target := "hello world"
	e := New(target)

	for _, char := range "hello" {
		e.ProcessKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}

	if string(e.UserInput) != "hello" {
		t.Errorf("Expected input 'hello', got '%s'", string(e.UserInput))
	}

	if e.IsFinished {
		t.Error("Engine should not be finished yet")
	}

	for _, char := range " world" {
		e.ProcessKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}

	if !e.IsFinished {
		t.Error("Engine should be finished")
	}
}

func TestEngineErrors(t *testing.T) {
	target := "hello"
	e := New(target)

	for _, char := range "hella" {
		e.ProcessKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}

	if e.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", e.ErrorCount)
	}
}

func TestDeleteLastWord(t *testing.T) {
	e := New("test")

	cases := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello "},
		{"hello  ", ""},
		{"hello", ""},
		{"", ""},
		{"func main()", "func "},
	}

	for _, c := range cases {
		result := e.deleteLastWord([]rune(c.input))
		if string(result) != c.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", c.input, c.expected, string(result))
		}
	}
}

func TestStatsCalculation(t *testing.T) {
	target := "abcde"
	e := New(target)

	e.StartTime = time.Now().Add(-60 * time.Second)
	e.EndTime = time.Now()
	e.IsFinished = true

	wpm, accuracy, _ := e.GetStats()

	if wpm < 0.9 || wpm > 1.1 {
		t.Errorf("Expected ~1 WPM for 5 chars in 60s, got %f", wpm)
	}

	if accuracy != 100.0 {
		t.Errorf("Expected 100%% accuracy, got %f", accuracy)
	}
}

func TestWPMWithErrors(t *testing.T) {
	target := "hello"
	e := New(target)

	for _, char := range "hella" {
		e.ProcessKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}

	e.StartTime = time.Now().Add(-60 * time.Second)
	e.EndTime = time.Now()

	wpm, _, _ := e.GetStats()

	expectedCorrect := 5 - 1
	expectedWords := float64(expectedCorrect) / 5.0
	expectedWPM := expectedWords * 1.0

	if wpm < expectedWPM-0.1 || wpm > expectedWPM+0.1 {
		t.Errorf("Expected ~%.2f WPM (4 correct chars / 5), got %.2f", expectedWPM, wpm)
	}
}

func TestAccuracyCalculation(t *testing.T) {
	cases := []struct {
		name        string
		target      string
		input       string
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "perfect accuracy",
			target:      "hello",
			input:       "hello",
			expectedMin: 99.9,
			expectedMax: 100.0,
		},
		{
			name:        "one error in five chars",
			target:      "hello",
			input:       "hella",
			expectedMin: 79.9,
			expectedMax: 80.1,
		},
		{
			name:        "multiple errors",
			target:      "hello",
			input:       "hxllo",
			expectedMin: 79.9,
			expectedMax: 80.1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := New(tc.target)
			for _, char := range tc.input {
				e.ProcessKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
			}

			_, accuracy, _ := e.GetStats()

			if accuracy < tc.expectedMin || accuracy > tc.expectedMax {
				t.Errorf("Expected accuracy between %.2f%% and %.2f%%, got %.2f%%",
					tc.expectedMin, tc.expectedMax, accuracy)
			}
		})
	}
}
