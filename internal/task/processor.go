package task

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Processor turns the input text of a task into a Result.
// Having an interface lets the service be tested with fake processors.
type Processor interface {
	Process(text string) (Result, error)
}

// TextAnalyzer is the real Processor: it computes text statistics.
type TextAnalyzer struct{}

// Process counts words, runes, lines and unique words in text.
func (TextAnalyzer) Process(text string) (Result, error) {
	if err := ValidateText(text); err != nil {
		return Result{}, err
	}

	var res Result
	res.Characters = utf8.RuneCountInString(text)

	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			res.Lines++
		}
	}

	seen := make(map[string]struct{})
	for _, field := range strings.Fields(text) {
		res.Words++

		word := strings.TrimFunc(field, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if word == "" {
			continue
		}
		seen[strings.ToLower(word)] = struct{}{}
		if utf8.RuneCountInString(word) > utf8.RuneCountInString(res.LongestWord) {
			res.LongestWord = word
		}
	}
	res.UniqueWords = len(seen)
	return res, nil
}
