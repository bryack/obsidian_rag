package tokenizer

import (
	"hash/fnv"
	"strings"
	"unicode"

	"github.com/bryack/obsidian_rag/internal/domain"
)

type Tokenizer struct {
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) ExtractTerms(text string) map[string]int {
	counts := make(map[string]int)

	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for _, word := range words {
		if len(word) < 2 {
			continue
		}

		if _, ok := stopWords[word]; ok {
			continue
		}
		counts[word]++
	}

	return counts
}

func (t *Tokenizer) ToBM25Vector(text string, stats *domain.BM25Stats) map[uint32]float32 {
	count := map[uint32]float32{}
	docFrequency := t.ExtractTerms(text)
	h := fnv.New32a()

	docLen := domain.SumTermFrequencies(docFrequency)

	for term := range docFrequency {
		h.Reset()
		TF := stats.CalculateTF(docFrequency[term], docLen)
		IDF := stats.CalculateIDF(term)
		h.Write([]byte(term))
		count[h.Sum32()] = float32(TF * IDF)
	}

	return count
}
