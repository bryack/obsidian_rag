package specifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type RAGProvider interface {
	Ask(question string) (string, error)
}

func RAGSpecification(t testing.TB, provider RAGProvider) {
	answer, err := provider.Ask("Что такое Obsidian RAG?")
	assert.NoError(t, err)
	assert.NotZero(t, answer)
}
