package domain

import "context"

type VectorStore interface {
	Save(ctx context.Context, doc Document) error
	Search(ctx context.Context, vector []float32, sparse map[uint32]float32) ([]Document, error)
	GetAllHashes(ctx context.Context) (map[string]string, error)
	SaveBatch(ctx context.Context, docs []Document) error
	SearchWithScope(ctx context.Context, query SearchQuery) ([]Document, error)
}

type SearchQuery struct {
	DenseVector  []float32
	SparseVector map[uint32]float32
	Scope        Scope
	Limit        int
}

type NoteRepository interface {
	GetNotes() ([]Document, error)
}

type Embedder interface {
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
	EmbedDocuments(ctx context.Context, text []string) ([][]float32, error)
}

type EmbeddingFormatter interface {
	Format(doc Document) string
}

type Tokenizer interface {
	ToSparseVector(text string) map[uint32]float32
}

// Scope represents a filtering criterion for search operations.
// Implementations define specific filtering strategies (folder, tags, etc.).
type Scope interface {
	Name() string
}

type Document struct {
	FilePath string `yaml:"file_path"`
	Hash     string `yaml:"hash"`

	Content    string   `yaml:"content"`
	HeaderPath []string `yaml:"header_path"`

	Metadata Metadata

	Vector VectorData
	Score  float32
}

type VectorData struct {
	Dense        []float32
	SparseVector map[uint32]float32
}

type Metadata struct {
	Tags    []string `yaml:"tags"`
	Project []string `yaml:"project"`
	Links   []string `yaml:"links"`
}
