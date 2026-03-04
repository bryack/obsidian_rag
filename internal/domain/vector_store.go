package domain

type VectorStore interface {
	Save(doc Document) error
	Search(vector []float32) ([]Document, error)
	GetAllHashes() (map[string]string, error)
	SaveBatch(docs []Document) error
}

type NoteRepository interface {
	GetNotes() ([]Document, error)
}

type Embedder interface {
	EmbedQuery(text string) ([]float32, error)
	EmbedDocuments(text []string) ([][]float32, error)
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
