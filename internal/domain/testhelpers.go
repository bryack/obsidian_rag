package domain

type SpyVectorStore struct {
	SaveCalled int
	Hashes     map[string]string
	Documents  []Document
}

func (s *SpyVectorStore) Save(doc Document) error {
	s.SaveCalled++
	s.Documents = append(s.Documents, doc)
	if s.Hashes == nil {
		s.Hashes = make(map[string]string)
	}
	s.Hashes[doc.FilePath] = doc.Hash
	return nil
}

func (s *SpyVectorStore) Search(vector []float32) ([]Document, error) {
	return s.Documents, nil
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

type StubParser struct {
	Items []Document
}

func (p *StubParser) Parse(doc Document) ([]Document, error) {
	return p.Items, nil
}

type SpyEmbedder struct {
	vector []float32
	Calls  [][]string
}

func (e *SpyEmbedder) EmbedQuery(text string) ([]float32, error) {
	return e.vector, nil
}

func (e *SpyEmbedder) EmbedDocuments(texts []string) ([][]float32, error) {
	e.Calls = append(e.Calls, texts)
	res := make([][]float32, len(texts))
	for i := range texts {
		res[i] = e.vector
	}
	return res, nil
}
