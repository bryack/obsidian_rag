package qdrant

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/google/uuid"
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

	ctx := context.Background()
	hashes := make(map[string]string)

	result, err := q.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: collectionName,
		WithPayload:    qdrant.NewWithPayload(true),
		Limit:          qdrant.PtrOf(uint32(10000)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scroll points: %w", err)
	}

	for _, p := range result {
		filepath := p.Payload["file_path"].GetStringValue()
		hash := p.Payload["hash"].GetStringValue()
		if filepath != "" {
			hashes[filepath] = hash
		}
	}
	return hashes, nil
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
	ctx := context.Background()
	pointID := uuid.New().String()

	waitUpsert := true
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewID(pointID),
				Vectors: qdrant.NewVectors(doc.Embedding...),
				Payload: map[string]*qdrant.Value{
					"file_path": qdrant.NewValueString(doc.FilePath),
					"hash":      qdrant.NewValueString(doc.Hash),
					"content":   qdrant.NewValueString(doc.Content),
				},
			},
		},
	})
	return err
}

func (q *QdrantStore) Search(vector []float32) ([]domain.Document, error) {
	ctx := context.Background()

	result, err := q.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(vector...),
		Limit:          qdrant.PtrOf(uint64(1)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed get result from query: %w", err)
	}

	var docs []domain.Document
	for _, p := range result {
		docs = append(docs, domain.Document{
			FilePath: p.Payload["file_path"].GetStringValue(),
			Hash:     p.Payload["hash"].GetStringValue(),
			Content:  p.Payload["content"].GetStringValue(),
		})
	}

	return docs, nil
}
