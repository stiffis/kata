package generator

import (
	"math/rand"
	"os"
	"strings"
	"time"
)

type LessonType int

const (
	TypeBigrams LessonType = iota
	TypeWords
	TypeSymbols
	TypeCode
	TypeFile
	TypeWeaknesses
)

type Generator struct {
	rand *rand.Rand
}

type WeakKey struct {
	Key       string
	ErrorRate float64
}

func New() *Generator {
	return &Generator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

var goBigrams = []string{
	"fu", "un", "nc", "ct", "ti", "io", "on", "re", "et", "tu", "ur", "rn",
	"if", "er", "rr", "ni", "il", "pa", "ac", "ck", "ka", "ag", "ge",
	"st", "tr", "ru", "uc", "in", "nt", "ty", "pe", "ra", "an", "ng",
}

var goKeywords = []string{
	"func", "package", "import", "return", "if", "else", "for", "range",
	"struct", "interface", "type", "var", "const", "go", "defer", "map",
	"chan", "select", "case", "switch", "break", "continue", "fallthrough",
	"goto", "true", "false", "nil", "error", "string", "int", "bool",
}

var goSnippets = []string{
	"func main() {\n\tfmt.Println(\"Hello\")\n}",
	"if err != nil {\n\treturn err\n}",
	"for i := 0; i < len(items); i++ {\n\tfmt.Println(items[i])\n}",
	"for _, item := range items {\n\tprocess(item)\n}",
	"type User struct {\n\tID   int\n\tName string\n}",
	"func (u *User) GetName() string {\n\treturn u.Name\n}",
	"ch := make(chan int, 10)",
	"defer file.Close()",
	"result, err := doSomething()",
	"switch value {\ncase 1:\n\treturn \"one\"\ndefault:\n\treturn \"other\"\n}",
}

var goSymbols = []string{
	"{}", "[]", "()", "!=", "==", "<=", ">=", "&&", "||", ":=", "...",
	"<-", "->", "*", "&", "%", "+=", "-=", "*=", "/=",
}

func (g *Generator) GenerateLesson(lessonType LessonType, length int) string {
	switch lessonType {
	case TypeBigrams:
		return g.generateFromList(goBigrams, length, " ")
	case TypeWords:
		return g.generateFromList(goKeywords, length, " ")
	case TypeSymbols:
		return g.generateFromList(goSymbols, length, " ")
	case TypeCode:
		return g.generateCode(length)
	default:
		return g.generateFromList(goKeywords, length, " ")
	}
}

func (g *Generator) GenerateFromFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (g *Generator) generateFromList(list []string, count int, sep string) string {
	var result []string
	for i := 0; i < count; i++ {
		result = append(result, list[g.rand.Intn(len(list))])
	}
	return strings.Join(result, sep)
}

func (g *Generator) generateCode(snippetCount int) string {
	var result []string
	for i := 0; i < snippetCount && i < len(goSnippets); i++ {
		result = append(result, goSnippets[g.rand.Intn(len(goSnippets))])
	}
	return strings.Join(result, "\n\n")
}

// GenerateWeaknessLesson creates a lesson focused on weak keys
func (g *Generator) GenerateWeaknessLesson(weakKeys []WeakKey, length int) string {
	if len(weakKeys) == 0 {
		// Fallback to keywords if no weakness data
		return g.generateFromList(goKeywords, length, " ")
	}

	// Build a pool of words containing weak keys
	var wordPool []string
	seen := make(map[string]bool) // Prevent duplicates

	// For each weak key, find words/bigrams that contain it
	for _, weak := range weakKeys {
		// Add keywords containing this character
		for _, word := range goKeywords {
			if strings.Contains(word, weak.Key) && !seen[word] {
				wordPool = append(wordPool, word)
				seen[word] = true
			}
		}

		// Add bigrams containing this character
		for _, bigram := range goBigrams {
			if strings.Contains(bigram, weak.Key) && !seen[bigram] {
				wordPool = append(wordPool, bigram)
				seen[bigram] = true
			}
		}
		
		// Add symbols containing this character
		for _, symbol := range goSymbols {
			if strings.Contains(symbol, weak.Key) && !seen[symbol] {
				wordPool = append(wordPool, symbol)
				seen[symbol] = true
			}
		}
	}

	// If pool is still empty or too small, add keywords
	if len(wordPool) == 0 {
		wordPool = append(wordPool, goKeywords...)
	}

	// Generate lesson
	var result []string
	for i := 0; i < length && len(wordPool) > 0; i++ {
		result = append(result, wordPool[g.rand.Intn(len(wordPool))])
	}

	// Ensure we have something
	if len(result) == 0 {
		return g.generateFromList(goKeywords, length, " ")
	}

	return strings.Join(result, " ")
}
