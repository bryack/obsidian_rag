package domain

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type RagEngine struct {
	repo           NoteRepository
	store          VectorStore
	parser         Parser
	embedder       Embedder
	tokenizer      Tokenizer
	formatter      EmbeddingFormatter
	contextBuilder ContextBuilder
	generator      AnswerGenerator
	batchSize      int
}

func NewRagEngine(repo NoteRepository, store VectorStore, parser Parser, tokenizer Tokenizer, embedder Embedder, formatter EmbeddingFormatter) *RagEngine {
	return &RagEngine{
		store:     store,
		repo:      repo,
		parser:    parser,
		embedder:  embedder,
		tokenizer: tokenizer,
		formatter: formatter,
		batchSize: 8,
	}
}

type AskQuery struct {
	Question string
	Scope    Scope
	Generate bool
}

func (re *RagEngine) Ask(ctx context.Context, query AskQuery) (string, error) {
	chunks, err := re.SearchChunks(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to find chunks: %w", err)
	}

	if !query.Generate {
		return re.formatSearchResults(query, chunks), nil
	}

	time.Sleep(200 * time.Millisecond)

	if re.generator == nil || re.contextBuilder == nil {
		return "", fmt.Errorf("generator or context builder not configured")
	}

	contextText := re.contextBuilder.BuildContext(chunks)

	answer, err := re.generator.Generate(ctx, query.Question, contextText)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	return answer, nil
}

func (re *RagEngine) formatSearchResults(query AskQuery, chunks []Document) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Результаты поиска для: %q (Область: %s)\n\n", query.Question, query.Scope.Name()))
	for i, chunk := range chunks {
		builder.WriteString(fmt.Sprintf("[%d] (Score: %.4f) Файл: %s\n", i+1, chunk.Score, chunk.FilePath))
		formatted := re.formatter.Format(chunk)
		builder.WriteString(formatted + "\n")
		builder.WriteString("------------------------------------------\n\n")
	}

	return builder.String()
}

func (re *RagEngine) SearchChunks(ctx context.Context, query AskQuery) ([]Document, error) {
	vector, err := re.embedder.EmbedQuery(ctx, query.Question)
	if err != nil {
		return nil, fmt.Errorf("failed to get vector for question %q: %w", query.Question, err)
	}

	sparse := re.tokenizer.ToSparseVector(query.Question)

	searchQuery := SearchQuery{
		DenseVector:  vector,
		SparseVector: sparse,
		Scope:        query.Scope,
		Limit:        10,
	}

	chunks, err := re.store.SearchWithScope(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search info for question: %q: %w", query.Question, err)
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("can't find any documents for question %q", query.Question)
	}
	return chunks, nil
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

	fmt.Fprintf(os.Stderr, "Debug: Found %d existing hashes in DB\n", len(hashes))

	var buffer []Document
	for i, doc := range docs {
		if re.needsSync(doc, hashes) {
			parcedChunks, err := re.prepareChunks(doc)
			if err != nil {
				return fmt.Errorf("failed to prepare chunks: %w", err)
			}
			buffer = append(buffer, parcedChunks...)
			if len(buffer) >= re.batchSize {
				if err := re.processBatch(ctx, buffer); err != nil {
					return fmt.Errorf("failed to process batch: %w", err)
				}
				buffer = buffer[:0]
			}
			if i%100 == 0 {
				fmt.Fprintf(os.Stderr, "[%d/%d] Indexed: %s\n", i+1, len(docs), doc.FilePath)
			}
		}
	}

	if len(buffer) > 0 {
		return re.processBatch(ctx, buffer)
	}

	fmt.Fprintf(os.Stderr, "Indexed %d notes\n", len(docs))
	return nil
}

func (re *RagEngine) needsSync(doc Document, existingHashes map[string]string) bool {
	existingHash, ok := existingHashes[doc.FilePath]
	return !ok || existingHash != doc.Hash
}

func (re *RagEngine) prepareChunks(doc Document) ([]Document, error) {
	parcedChunks, err := re.parser.Parse(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse doc %q: %w", doc.FilePath, err)
	}
	if len(parcedChunks) == 0 {
		parcedChunks = append(parcedChunks, Document{
			FilePath: doc.FilePath,
			Hash:     doc.Hash,
			Metadata: doc.Metadata,
			Content:  "",
		})
	}
	return parcedChunks, nil
}

func (re *RagEngine) processBatch(ctx context.Context, batch []Document) error {
	var textToEmbed []string
	var indicesToEmbed []int

	for i, c := range batch {
		content := strings.TrimSpace(c.Content)
		if content == "" {
			batch[i].Vector.Dense = make([]float32, 1024)
		} else {
			batch[i].Vector.SparseVector = re.tokenizer.ToSparseVector(content)
			formatted := re.formatter.Format(c)
			textToEmbed = append(textToEmbed, formatted)
			indicesToEmbed = append(indicesToEmbed, i)
		}
	}

	if len(textToEmbed) > 0 {
		vectors, err := re.embedder.EmbedDocuments(ctx, textToEmbed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG Error: failed to embed batch of %d texts. First text length: %d\n", len(textToEmbed), len(textToEmbed[0]))
			return fmt.Errorf("failed to embed chunk content: %w", err)
		}

		for j, vector := range vectors {
			batch[indicesToEmbed[j]].Vector.Dense = vector
		}
	}

	if err := re.store.SaveBatch(ctx, batch); err != nil {
		return fmt.Errorf("failed to save batch: %w", err)
	}
	return nil
}

func (re *RagEngine) SetGenerator(generator AnswerGenerator, contextBuilder ContextBuilder) {
	re.generator = generator
	re.contextBuilder = contextBuilder
}
