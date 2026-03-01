package domain

type SpyVectorStore struct {
	SaveCalled int
	Hashes     map[string]string
}

func (s *SpyVectorStore) Save(doc Document) error {
	s.SaveCalled++
	if s.Hashes == nil {
		s.Hashes = make(map[string]string)
	}
	s.Hashes[doc.FilePath] = doc.Hash
	return nil
}

func (s *SpyVectorStore) Search(query string) ([]Document, error) {
	return []Document{}, nil
}

func (s *SpyVectorStore) GetAllHashes() (map[string]string, error) {
	return s.Hashes, nil
}

type StubNoteRepository struct {
	Doc Document
}

func (s *StubNoteRepository) GetNotes() ([]Document, error) {
	return []Document{s.Doc}, nil
}
