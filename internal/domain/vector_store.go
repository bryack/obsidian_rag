package domain

type VectorStore interface {
	Save(doc Document) error
	Search(query string) ([]Document, error)
	GetAllHashes() (map[string]string, error)
}

type NoteRepository interface {
	GetNotes() ([]Document, error)
}

type Document struct {
	FilePath string
	Hash     string
	Content  string
	Metadata map[string]interface{}
}
