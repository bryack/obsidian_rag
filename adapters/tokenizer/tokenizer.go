package tokenizer

import (
	"hash/fnv"
	"strings"
	"unicode"
)

type Tokenizer struct {
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) ToSparseVector(text string) map[uint32]float32 {
	counts := make(map[uint32]float32)

	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for _, word := range words {
		word := strings.Trim(word, ".,!?-()[]{}'\"")
		if len(word) < 2 {
			continue
		}

		lower := strings.ToLower(word)
		if _, ok := stopWords[lower]; ok {
			continue
		}

		h := fnv.New32a()
		h.Write([]byte(word))
		counts[h.Sum32()]++
	}

	return counts
}
