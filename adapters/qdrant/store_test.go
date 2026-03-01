package qdrant

import (
	"context"
	"testing"

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
}
