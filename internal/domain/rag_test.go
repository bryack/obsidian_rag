package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type SpyVectorStore struct {
	saveCalled int
	hashes     map[string]string
}

func (s *SpyVectorStore) Save(doc Document) error {
	s.saveCalled++
	if s.hashes == nil {
		s.hashes = make(map[string]string)
	}
	s.hashes[doc.FilePath] = doc.Hash
	return nil
}

func (s *SpyVectorStore) Search(query string) (Chunks, error) {
	return Chunks{}, nil
}

func (s *SpyVectorStore) GetAllHashes() (map[string]string, error) {
	return s.hashes, nil
}

type StubNoteRepository struct {
	doc Document
}

func (s *StubNoteRepository) GetNotes(dirPath string) (Chunks, error) {
	return Chunks{s.doc}, nil
}

func TestRagEngine_Ask(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{}
		engine := NewRagEngine(repo, store)
		answer, err := engine.Ask("Что такое Obsidian RAG?")

		assert.NoError(t, err)
		assert.Equal(t, "Obsidian RAG: Ответ найден в ваших заметках.", answer)
	})
}

func TestRagEngine_Sync(t *testing.T) {
	store := &SpyVectorStore{}
	repo := &StubNoteRepository{}
	engine := NewRagEngine(repo, store)

	err := engine.Sync("any-path")
	assert.NoError(t, err)
	assert.Equal(t, 1, store.saveCalled)
	assert.True(t, len(store.hashes) > 0)

	err = engine.Sync("any-path")
	assert.NoError(t, err)
	assert.Equal(t, 1, store.saveCalled)

}
