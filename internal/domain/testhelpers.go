package domain

import (
	"context"
)

type SpyVectorStore struct {
	SaveCalled      int
	Hashes          map[string]string
	Documents       []Document
	LastSearchScope Scope
	DeletedPaths    []string
}

func (s *SpyVectorStore) Save(ctx context.Context, doc Document) error {
	s.SaveCalled++
	s.Documents = append(s.Documents, doc)
	if s.Hashes == nil {
		s.Hashes = make(map[string]string)
	}
	s.Hashes[doc.FilePath] = doc.Hash
	return nil
}

func (s *SpyVectorStore) Search(ctx context.Context, vector []float32, sparse map[uint32]float32) ([]Document, error) {
	return s.Documents, nil
}

func (s *SpyVectorStore) GetAllHashes(ctx context.Context) (map[string]string, error) {
	return s.Hashes, nil
}

func (s *SpyVectorStore) SaveBatch(ctx context.Context, docs []Document) error {
	for _, doc := range docs {
		if err := s.Save(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *SpyVectorStore) SearchWithScope(ctx context.Context, query SearchQuery) ([]Document, error) {
	s.LastSearchScope = query.Scope
	return s.Documents, nil
}

func (s *SpyVectorStore) DeleteByFilePaths(ctx context.Context, filePaths []string) error {
	s.DeletedPaths = append(s.DeletedPaths, filePaths...)
	for _, path := range filePaths {
		delete(s.Hashes, path)
	}
	return nil
}

type StubNoteRepository struct {
	Docs []Document
}

func (s *StubNoteRepository) GetNotes() ([]Document, error) {
	return s.Docs, nil
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

func (e *SpyEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.vector, nil
}

func (e *SpyEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	e.Calls = append(e.Calls, texts)
	res := make([][]float32, len(texts))
	for i := range texts {
		res[i] = e.vector
	}
	return res, nil
}

type StubTokenizer struct{}

func (st *StubTokenizer) ExtractTerms(text string) map[string]int {
	return map[string]int{"test": 1}
}

func (st *StubTokenizer) ToBM25Vector(text string, stats *BM25Stats) map[uint32]float32 {
	return map[uint32]float32{1: 1.0}
}

type SpyGenerator struct {
	Answer string
}

func (g *SpyGenerator) Generate(ctx context.Context, question string, context string) (string, error) {
	return g.Answer, nil
}

type StubContextBuilder struct{}

func (cb *StubContextBuilder) BuildContext(chunks []Document) string {
	return ""
}
