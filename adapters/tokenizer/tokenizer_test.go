package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizer_ToSparseVector(t *testing.T) {
	tokenizer := NewTokenizer()
	text := "North-Star"

	sparse := tokenizer.ExtractTerms(text)

	assert.Len(t, sparse, 2)
	for _, weight := range sparse {
		assert.True(t, weight > 0)
	}
}

func TestTokenizer_StopWords(t *testing.T) {
	tokenizer := NewTokenizer()

	t.Run("ignores stop words", func(t *testing.T) {
		text := "и, к, из, РЕДКО, the, and в Обсидиан"
		vector := tokenizer.ExtractTerms(text)

		assert.Equal(t, 1, len(vector))
	})
}
