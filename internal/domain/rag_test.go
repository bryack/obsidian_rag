package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsk(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		rag := NewRagEngine()
		answer, err := rag.Ask("Что такое Obsidian RAG?")

		assert.NoError(t, err)
		assert.Equal(t, "Obsidian RAG: Ответ найден в ваших заметках.", answer)
	})
}
