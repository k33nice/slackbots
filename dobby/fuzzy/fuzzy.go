package fuzzy

import (
	"regexp"

	"github.com/sajari/fuzzy"
)

// Fuzzy - typo autocorector.
type Fuzzy struct {
	model *fuzzy.Model

	// Dict - dictionary for model training.
	Dict []string
}

// SetUp - prepare fuzzy typo fixer
func (f *Fuzzy) SetUp() {
	f.model = fuzzy.NewModel()

	f.model.SetThreshold(1)
	f.model.SetDepth(5)
	f.model.Train(f.Dict)
	f.model.TrainWord("single")
}

// Fix - correct typo in passed word relay on dictionary.
func (f *Fuzzy) Fix(word string) string {
	correct := f.model.SpellCheck(word)
	re := regexp.MustCompile(`[^A-Za-z0-9]`)
	if correct != "" {
		return correct
	}
	return re.ReplaceAllString(word, "")
}
