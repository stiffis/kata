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
	target := "abcde" // 5 chars = 1 standard word
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

