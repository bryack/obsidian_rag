package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRagEngine_Ask(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{}
		engine := NewRagEngine(repo, store)
		answer, err := engine.Ask("Что такое Obsidian RAG?")

		assert.NoError(t, err)
		assert.Equal(t, "Obsidian RAG: Ответ найден в ваших заметках.", answer)
	})
}

func TestRagEngine_Sync(t *testing.T) {
	store := &SpyVectorStore{}
	repo := &StubNoteRepository{
		Doc: Document{FilePath: "note.md", Hash: "v1"},
	}
	engine := NewRagEngine(repo, store)

	err := engine.Sync()
	assert.NoError(t, err)
	assert.Equal(t, 1, store.SaveCalled)
	assert.Equal(t, "v1", store.Hashes["note.md"])

	err = engine.Sync()
	assert.NoError(t, err)
	assert.Equal(t, 1, store.SaveCalled)

	repo.Doc.Hash = "v2"
	err = engine.Sync()
	assert.NoError(t, err)
	assert.Equal(t, 2, store.SaveCalled)
	assert.Equal(t, "v2", store.Hashes["note.md"])
}
