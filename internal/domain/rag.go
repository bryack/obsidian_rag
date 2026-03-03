package domain

import (
	"fmt"
)

type RagEngine struct {
	repo     NoteRepository
	store    VectorStore
	parser   Parser
	embedder Embedder
}

func NewRagEngine(repo NoteRepository, store VectorStore, parser Parser, embedder Embedder) *RagEngine {
	return &RagEngine{store: store, repo: repo, parser: parser, embedder: embedder}
}

func (re *RagEngine) Ask(question string) (string, error) {
	chunks, err := re.store.Search(question)
	if err != nil {
		return "", fmt.Errorf("failed to search info for question: %q: %w", question, err)
	}

	if len(chunks) == 0 {
		return "", fmt.Errorf("can't find any documents for question %q", question)
	}

	return chunks[0].Content, nil
}

func (re *RagEngine) Sync() error {
	hashes, err := re.store.GetAllHashes()
	if err != nil {
		return fmt.Errorf("failed to get hashes: %w", err)
	}

	docs, err := re.repo.GetNotes()
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	for _, doc := range docs {
		existingHash, ok := hashes[doc.FilePath]
		if !ok || existingHash != doc.Hash {
			parcedChunks, err := re.parser.Parse(doc)
			if err != nil {
				return fmt.Errorf("failed to parse doc %q: %w", doc.FilePath, err)
			}
			for _, chunk := range parcedChunks {
				err := re.store.Save(chunk)
				if err != nil {
					return fmt.Errorf("failed to save document: %w", err)
				}
			}
		}
	}
	fmt.Printf("Indexed %d notes\n", len(docs))
	return nil
}
