package qdrant

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/qdrant/go-client/qdrant"
)

const (
	collectionName = "obsidian_notes"
	VectorSize     = 1024
)

type QdrantStore struct {
	client *qdrant.Client
}

func NewQdrantStore(grpcEndpoint string) (*QdrantStore, error) {
	host, portStr, err := net.SplitHostPort(grpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid grpc endpoint: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q in endpoint: %w", portStr, err)
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	return &QdrantStore{client: client}, nil
}

func (q *QdrantStore) GetAllHashes() (map[string]string, error) {
	if err := q.ensureCollection(); err != nil {
		return nil, err
	}
	return map[string]string{}, nil
}

func (q *QdrantStore) ensureCollection() error {
	ctx := context.Background()

	exists, err := q.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		err := q.client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}
	return nil
}

func (q *QdrantStore) Save(doc domain.Document) error {
	return fmt.Errorf("not implemented")
}

func (q *QdrantStore) Search(query string) ([]domain.Document, error) {
	return []domain.Document{}, fmt.Errorf("not implemented")
}
