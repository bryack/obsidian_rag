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
	vector, err := re.embedder.Embed(question)
	if err != nil {
		return "", fmt.Errorf("failed to get vector for question %q: %w", question, err)
	}
	chunks, err := re.store.Search(vector)
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

	fmt.Printf("Debug: Found %d existing hashes in DB\n", len(hashes))
	for i, doc := range docs {
		existingHash, ok := hashes[doc.FilePath]
		if !ok || existingHash != doc.Hash {
			fmt.Printf("Debug: File %s not found in DB hashes\n", doc.FilePath)
			parcedChunks, err := re.parser.Parse(doc)
			if err != nil {
				return fmt.Errorf("failed to parse doc %q: %w", doc.FilePath, err)
			}
			if len(parcedChunks) == 0 {
				parcedChunks = append(parcedChunks, Document{
					FilePath: doc.FilePath,
					Hash:     doc.Hash,
					Metadata: doc.Metadata,
					Content:  "",
				})
			}
			fmt.Printf("[%d/%d] Indexed: %s\n", i+1, len(docs), doc.FilePath)

			for _, chunk := range parcedChunks {
				vector, err := re.embedder.Embed(chunk.Content)
				if err != nil {
					return fmt.Errorf("failed to embed chunk content for file %q: %w", doc.FilePath, err)
				}
				chunk.Embedding = vector

				if err = re.store.Save(chunk); err != nil {
					return fmt.Errorf("failed to save document: %w", err)
				}
			}
		}
	}
	fmt.Printf("Indexed %d notes\n", len(docs))
	return nil
}
