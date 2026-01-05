package generator

import (
	"strings"
	"testing"
)

func TestGeneratorLanguages(t *testing.T) {
	g := New()
	
	langs := []Language{LangGo, LangEnglish, LangSpanish, LangPython, LangRust}
	
	for _, lang := range langs {
		g.SetLanguage(lang)
		lesson := g.GenerateLesson(TypeWords, 5)
		
		words := strings.Split(lesson, " ")
		if len(words) != 5 {
			t.Errorf("Language %s: expected 5 words, got %d", lang, len(words))
		}
		
		if lesson == "" {
			t.Errorf("Language %s: generated empty lesson", lang)
		}
	}
}

func TestGenerateCode(t *testing.T) {
	g := New()
	g.SetLanguage(LangGo)
	
	lesson := g.GenerateLesson(TypeCode, 2)
	
	// Code snippets are joined by \n\n
	if !strings.Contains(lesson, "\n") && !strings.Contains(lesson, "func") {
		t.Error("Generated code doesn't look like Go code")
	}
}

func TestWeaknessLessonFallback(t *testing.T) {
	g := New()
	// Test with no weak keys (should fallback to words)
	lesson := g.GenerateWeaknessLesson([]WeakKey{}, 5)
	
	words := strings.Split(lesson, " ")
	if len(words) != 5 {
		t.Errorf("Expected 5 words in fallback, got %d", len(words))
	}
}
