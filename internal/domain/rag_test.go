package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRagEngine_Ask(t *testing.T) {

	t.Run("real search", func(t *testing.T) {
		store := &SpyVectorStore{
			Documents: []Document{
				{Content: "В Obsidian RAG используется Go."},
			},
		}
		repo := &StubNoteRepository{}
		parser := &StubParser{}
		embedder := &StubEmbedder{}

		engine := NewRagEngine(repo, store, parser, embedder)
		engine.Sync()

		answer, err := engine.Ask("На чем написан проект?")

		assert.NoError(t, err)
		assert.Contains(t, answer, "Go")
	})
}

func TestRagEngine_Sync(t *testing.T) {
	store := &SpyVectorStore{}
	repo := &StubNoteRepository{
		Doc: Document{FilePath: "note.md", Hash: "v1"},
	}
	parser := &StubParser{}
	embedder := &StubEmbedder{}

	engine := NewRagEngine(repo, store, parser, embedder)

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
