package qdrant

import (
	"fmt"

	"github.com/bryack/obsidian_rag/internal/domain"
)

type QdrantStore struct{}

func NewQdrantStore(grpcEndpoint string) *QdrantStore {
	return &QdrantStore{}
}

func (q *QdrantStore) GetAllHashes() (map[string]string, error) {
	return map[string]string{}, fmt.Errorf("not implemented")
}

func (q *QdrantStore) Save(doc domain.Document) error {
	return fmt.Errorf("not implemented")
}

func (q *QdrantStore) Search(query string) ([]domain.Document, error) {
	return []domain.Document{}, fmt.Errorf("not implemented")
}
