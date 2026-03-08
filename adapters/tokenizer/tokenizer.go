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

	h := fnv.New32a()

	for _, word := range words {
		if len(word) < 2 {
			continue
		}

		if _, ok := stopWords[word]; ok {
			continue
		}

		h.Write([]byte(word))
		counts[h.Sum32()]++
	}

	return counts
}
