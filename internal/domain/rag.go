package domain

import (
	"context"
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

func (re *RagEngine) Ask(ctx context.Context, question string) (string, error) {
	vector, err := re.embedder.EmbedQuery(ctx, question)
	if err != nil {
		return "", fmt.Errorf("failed to get vector for question %q: %w", question, err)
	}
	chunks, err := re.store.Search(ctx, vector)
	if err != nil {
		return "", fmt.Errorf("failed to search info for question: %q: %w", question, err)
	}

	if len(chunks) == 0 {
		return "", fmt.Errorf("can't find any documents for question %q", question)
	}

	return chunks[0].Content, nil
}

func (re *RagEngine) Sync(ctx context.Context) error {
	hashes, err := re.store.GetAllHashes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get hashes: %w", err)
	}

	docs, err := re.repo.GetNotes()
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	fmt.Printf("Debug: Found %d existing hashes in DB\n", len(hashes))

	var allNewChunks []Document
	for i, doc := range docs {
		existingHash, ok := hashes[doc.FilePath]
		if !ok || existingHash != doc.Hash {
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
			allNewChunks = append(allNewChunks, parcedChunks...)
			fmt.Printf("[%d/%d] Indexed: %s\n", i+1, len(docs), doc.FilePath)
		}
	}

	fmt.Printf("Debug: Total new chunks to process: %d\n", len(allNewChunks))
	if len(allNewChunks) == 0 {
		fmt.Println("Debug: No new chunks found. Skipping batching.")
		return nil
	}

	batchSize := 32
	for i := 0; i < len(allNewChunks); i += batchSize {
		end := i + batchSize
		if end > len(allNewChunks) {
			end = len(allNewChunks)
		}

		batch := allNewChunks[i:end]

		var textToEmbed []string
		for _, c := range batch {
			textToEmbed = append(textToEmbed, c.Content)
		}

		vectors, err := re.embedder.EmbedDocuments(ctx, textToEmbed)
		if err != nil {
			fmt.Printf("DEBUG Error: failed to embed batch of %d texts. First text length: %d\n", len(textToEmbed), len(textToEmbed[0]))
			return fmt.Errorf("failed to embed chunk content: %w", err)
		}

		for j := range batch {
			batch[j].Embedding = vectors[j]
		}

		if err = re.store.SaveBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to save document: %w", err)
		}
	}
	fmt.Printf("Indexed %d notes\n", len(docs))
	return nil
}
