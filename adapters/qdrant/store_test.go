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

	hashes, err := store.GetAllHashes()
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

	err = store.Save(doc)
	require.NoError(t, err)

	t.Run("Search returns Document", func(t *testing.T) {
		result, err := store.Search(testVector)
		require.NoError(t, err)

		require.NotEmpty(t, result)
		assert.Equal(t, "note.md", result[0].FilePath)
		assert.Equal(t, "В Obsidian RAG используется Go.", result[0].Content)
	})

	t.Run("GetAllHashes returns saved hashes", func(t *testing.T) {
		hashes, err := store.GetAllHashes()
		require.NoError(t, err)
		assert.Equal(t, doc.Hash, hashes["note.md"])
	})
}
