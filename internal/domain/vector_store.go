package domain

import "context"

type VectorStore interface {
	Save(ctx context.Context, doc Document) error
	Search(ctx context.Context, vector []float32) ([]Document, error)
	GetAllHashes(ctx context.Context) (map[string]string, error)
	SaveBatch(ctx context.Context, docs []Document) error
}

type NoteRepository interface {
	GetNotes() ([]Document, error)
}

type Embedder interface {
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
	EmbedDocuments(ctx context.Context, text []string) ([][]float32, error)
}

type Document struct {
	FilePath  string
	Hash      string
	Content   string
	Metadata  Metadata
	Embedding []float32
}

type Metadata struct {
	Tags    []string `yaml:"tags"`
	Project []string `yaml:"project"`
}
