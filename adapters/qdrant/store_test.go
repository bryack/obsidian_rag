package qdrant

import (
	"context"
	"testing"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func TestQdrant_Integration(t *testing.T) {
	ctx := context.Background()

	container, err := qdrant.Run(ctx, "qdrant/qdrant:latest")
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	grpcEndpoint, err := container.GRPCEndpoint(ctx)
	require.NoError(t, err)

	store, err := NewQdrantStore(grpcEndpoint)
	require.NoError(t, err)

	hashes, err := store.GetAllHashes(ctx)
	require.NoError(t, err)
	require.NotNil(t, hashes)

	testVector := make([]float32, 1024)
	testVector[0] = 1.0

	doc := domain.Document{
		FilePath:  "note.md",
		Hash:      "hash-of-file",
		Content:   "В Obsidian RAG используется Go.",
		Embedding: testVector,
	}

	err = store.Save(ctx, doc)
	require.NoError(t, err)

	t.Run("Search returns Document", func(t *testing.T) {
		result, err := store.Search(ctx, testVector)
		require.NoError(t, err)

		require.NotEmpty(t, result)
		assert.Equal(t, "note.md", result[0].FilePath)
		assert.Equal(t, "В Obsidian RAG используется Go.", result[0].Content)
	})

	t.Run("GetAllHashes returns saved hashes", func(t *testing.T) {
		hashes, err := store.GetAllHashes(ctx)
		require.NoError(t, err)
		assert.Equal(t, doc.Hash, hashes["note.md"])
	})
	t.Run("double Save does not increase the number of points", func(t *testing.T) {
		initialCount, err := store.CountPoints(ctx)
		require.NoError(t, err)

		err = store.Save(ctx, doc)
		require.NoError(t, err)

		finalCount, err := store.CountPoints(ctx)
		require.NoError(t, err)

		assert.Equal(t, initialCount, finalCount)
	})

	t.Run("save batch", func(t *testing.T) {
		err := store.clear(ctx)
		require.NoError(t, err)

		testVector1 := make([]float32, 1024)
		testVector1[1] = 1.0
		testVector2 := make([]float32, 1024)
		testVector2[2] = 1.0
		testVector3 := make([]float32, 1024)
		testVector3[3] = 1.0
		docs := []domain.Document{
			{FilePath: "note1.md", Hash: "hash-of-file1", Content: "В Obsidian RAG используется Ollama.", Embedding: testVector1},
			{FilePath: "note2.md", Hash: "hash-of-file2", Content: "В Obsidian RAG используется Qdrant.", Embedding: testVector2},
			{FilePath: "note3.md", Hash: "hash-of-file3", Content: "В Obsidian RAG используется Goldmark.", Embedding: testVector3},
		}

		err = store.SaveBatch(ctx, docs)
		assert.NoError(t, err)

		for _, v := range docs {
			result, err := store.Search(ctx, v.Embedding)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Equal(t, v.FilePath, result[0].FilePath, "Should find correct file for vector")
			assert.Equal(t, v.Content, result[0].Content)
		}
	})

	t.Run("should not find any empty notes", func(t *testing.T) {
		err := store.clear(ctx)
		require.NoError(t, err)

		testVector1 := make([]float32, 1024)
		testVector1[1] = 0.9
		testVector2 := make([]float32, 1024)
		testVector2[1] = 0.9
		docs := []domain.Document{
			{FilePath: "doc_with_content.md", Hash: "hash-of-file1", Content: "В Obsidian RAG используется Ollama. Ollama — это популярный бесплатный инструмент с открытым исходным кодом", Embedding: testVector1},
			{FilePath: "empty.md", Hash: "hash-of-file2", Content: "", Embedding: testVector2},
		}

		err = store.SaveBatch(ctx, docs)
		assert.NoError(t, err)

		searchVector := make([]float32, 1024)
		searchVector[1] = 0.9
		result, err := store.Search(ctx, searchVector)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		assert.Equal(t, 1, len(result), "Should find only 1 file for vector")
		assert.Equal(t, "doc_with_content.md", result[0].FilePath)
		assert.NotContains(t, "empty.md", result)
	})
}
