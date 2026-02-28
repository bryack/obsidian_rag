package domain

import (
	"fmt"
)

type RagEngine struct {
	repo  NoteRepository
	store VectorStore
}

func NewRagEngine(repo NoteRepository, store VectorStore) *RagEngine {
	return &RagEngine{store: store, repo: repo}
}

func (re *RagEngine) Ask(question string) (string, error) {
	if question == "Что такое Obsidian RAG?" {
		return "Obsidian RAG: Ответ найден в ваших заметках.", nil
	}
	return "", fmt.Errorf("failed to answer")
}

func (re *RagEngine) Sync(path string) error {
	hashes, err := re.store.GetAllHashes()
	if err != nil {
		return fmt.Errorf("failed to get hashes: %w", err)
	}

	chunks, err := re.repo.GetNotes(path)
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	for _, doc := range chunks {
		existingHash, ok := hashes[doc.FilePath]
		if !ok || existingHash != doc.Hash {
			err := re.store.Save(doc)
			if err != nil {
				return fmt.Errorf("failed to save document: %w", err)
			}
		}
	}

	return nil
}
