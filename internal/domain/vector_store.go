package domain

type VectorStore interface {
	Save(doc Document) error
	Search(query string) (Chunks, error)
	GetAllHashes() (map[string]string, error)
}

type NoteRepository interface {
	GetNotes() (Chunks, error)
}

type Document struct {
	FilePath string
	Hash     string
	Content  string
	Metadata map[string]interface{}
}

type Chunks []Document
