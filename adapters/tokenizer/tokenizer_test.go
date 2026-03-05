package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizer_ToSparseVector(t *testing.T) {
	tokenizer := NewTokenizer()
	text := "North-Star"

	sparse := tokenizer.ToSparseVector(text)

	assert.Len(t, sparse, 2)
	for _, weight := range sparse {
		assert.True(t, weight > 0)
	}
}
