package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		embedder := &SpyEmbedder{}

		engine := NewRagEngine(repo, store, parser, embedder)
		engine.Sync()

		answer, err := engine.Ask("На чем написан проект?")

		assert.NoError(t, err)
		assert.Contains(t, answer, "Go")
	})
}

func TestRagEngine_Sync(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{
			Docs: []Document{{FilePath: "note.md", Hash: "v1", Content: "Hello!"}},
		}
		parser := &StubParser{Items: []Document{{FilePath: "note.md", Hash: "v1", Content: "Hello!"}}}
		embedder := &SpyEmbedder{vector: []float32{0.1, 0.2}}

		engine := NewRagEngine(repo, store, parser, embedder)

		err := engine.Sync()
		assert.NoError(t, err)
		assert.Equal(t, 1, store.SaveCalled)
		assert.Equal(t, "v1", store.Hashes["note.md"])
		require.Len(t, store.Documents, 1)
		assert.Equal(t, []float32{0.1, 0.2}, store.Documents[0].Embedding, "Document should be embedded before saving")

		err = engine.Sync()
		assert.NoError(t, err)
		assert.Equal(t, 1, store.SaveCalled)

		repo.Docs[0].Hash = "v2"
		parser.Items[0].Hash = "v2"
		err = engine.Sync()
		assert.NoError(t, err)
		assert.Equal(t, 2, store.SaveCalled)
		assert.Equal(t, "v2", store.Hashes["note.md"])
	})
	t.Run("empty file", func(t *testing.T) {
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{
			Docs: []Document{{FilePath: "document.md", Hash: "d1", Content: ""}},
		}
		parser := &StubParser{}
		embedder := &SpyEmbedder{vector: []float32{0.1, 0.2}}

		engine := NewRagEngine(repo, store, parser, embedder)

		err := engine.Sync()
		assert.NoError(t, err)
		assert.Equal(t, 1, store.SaveCalled)
	})
	t.Run("batch processing", func(t *testing.T) {
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{Docs: []Document{
			{FilePath: "batch1.md", Hash: "b1"},
			{FilePath: "batch2.md", Hash: "b2"},
		}}
		parser := &StubParser{Items: []Document{
			{FilePath: "batch.md", Content: "chunk 1"},
			{FilePath: "batch.md", Content: "chunk 2"},
			{FilePath: "batch.md", Content: "chunk 3"},
		}}
		embedder := &SpyEmbedder{vector: []float32{0.1}}

		engine := NewRagEngine(repo, store, parser, embedder)
		err := engine.Sync()
		assert.NoError(t, err)

		assert.Equal(t, 1, len(embedder.Calls), "Expected only 1 call to EmbedDocuments")
		assert.Equal(t, 6, len(embedder.Calls[0]), "Expected 6 chunks total in the batch")
	})
}
