package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRagEngine_Ask(t *testing.T) {

	t.Run("real search", func(t *testing.T) {
		ctx := context.Background()
		formatter := &DefaultFormatter{}
		tokenizer := &StubTokenizer{}
		store := &SpyVectorStore{
			Documents: []Document{
				{Content: "В Obsidian RAG используется Go."},
			},
		}
		repo := &StubNoteRepository{}
		parser := &StubParser{}
		embedder := &SpyEmbedder{}

		engine := NewRagEngine(repo, store, parser, tokenizer, embedder, formatter)
		engine.Sync(ctx)

		query := AskQuery{
			Question: "На чем написан проект?",
			Scope:    AllScope{},
		}
		answer, err := engine.Ask(ctx, query)

		assert.NoError(t, err)
		assert.Contains(t, answer, "Go")
	})
}

func TestRagEngine_Sync(t *testing.T) {
	formatter := &DefaultFormatter{}
	t.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		tokenizer := &StubTokenizer{}
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{
			Docs: []Document{{FilePath: "note.md", Hash: "v1", Content: "Hello!"}},
		}
		parser := &StubParser{Items: []Document{{FilePath: "note.md", Hash: "v1", Content: "Hello!"}}}
		embedder := &SpyEmbedder{vector: []float32{0.1, 0.2}}

		engine := NewRagEngine(repo, store, parser, tokenizer, embedder, formatter)

		err := engine.Sync(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, store.SaveCalled)
		assert.Equal(t, "v1", store.Hashes["note.md"])
		require.Len(t, store.Documents, 1)
		assert.Equal(t, []float32{0.1, 0.2}, store.Documents[0].Vector.Dense, "Document should be embedded before saving")

		err = engine.Sync(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, store.SaveCalled)

		repo.Docs[0].Hash = "v2"
		parser.Items[0].Hash = "v2"
		err = engine.Sync(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, store.SaveCalled)
		assert.Equal(t, "v2", store.Hashes["note.md"])
	})
	t.Run("empty file", func(t *testing.T) {
		tokenizer := &StubTokenizer{}
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{
			Docs: []Document{{FilePath: "document.md", Hash: "d1", Content: ""}},
		}
		parser := &StubParser{}
		embedder := &SpyEmbedder{vector: []float32{0.1, 0.2}}

		engine := NewRagEngine(repo, store, parser, tokenizer, embedder, formatter)

		err := engine.Sync(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, store.SaveCalled)
	})
	t.Run("batch processing", func(t *testing.T) {
		tokenizer := &StubTokenizer{}
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{Docs: []Document{
			{FilePath: "batch1.md", Hash: "b1"},
			{FilePath: "batch2.md", Hash: "b2"},
		}}
		parser := &StubParser{Items: []Document{
			{FilePath: "batch.md", Content: "chunk 1"},
			{FilePath: "batch.md", Content: "chunk 2"},
			{FilePath: "batch.md", Content: "chunk 3"},
		}}
		embedder := &SpyEmbedder{vector: []float32{0.1}}

		engine := NewRagEngine(repo, store, parser, tokenizer, embedder, formatter)
		err := engine.Sync(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, 1, len(embedder.Calls), "Expected only 1 call to EmbedDocuments")
		assert.Equal(t, 6, len(embedder.Calls[0]), "Expected 6 chunks total in the batch")
	})

	t.Run("skips empty content", func(t *testing.T) {
		tokenizer := &StubTokenizer{}
		store := &SpyVectorStore{}
		repo := &StubNoteRepository{Docs: []Document{
			{FilePath: "real.md", Hash: "r1"},
			{FilePath: "empty.md", Hash: "e2"},
		}}
		parser := &StubParser{Items: []Document{
			{FilePath: "real.md", Content: "Hello World. And some text to be a good chunk to parse"},
			{FilePath: "empty.md", Content: ""},
		}}
		embedder := &SpyEmbedder{vector: []float32{0.1}}

		engine := NewRagEngine(repo, store, parser, tokenizer, embedder, formatter)
		err := engine.Sync(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, 1, len(embedder.Calls), "Expected only 1 call to Ollama")
		assert.Equal(t, 2, len(embedder.Calls[0]), "Expected 2 chunks total in the batch")
		assert.Equal(t, 4, len(store.Documents), "Expected all docs saves for hashes")
		assert.Equal(t, 1, len(store.Documents[0].Vector.Dense))
		assert.Equal(t, 1024, len(store.Documents[1].Vector.Dense))
	})
}

func TestRagEngine_AskWithScope(t *testing.T) {
	ctx := context.Background()

	tokenizer := &StubTokenizer{}
	store := &SpyVectorStore{
		Documents: []Document{
			{FilePath: "work/note.md", Content: "test content", Score: 0.9},
		},
	}
	repo := &StubNoteRepository{
		Docs: []Document{
			{FilePath: "work/note.md", Hash: "v1", Content: "test"},
			{FilePath: "personal/note.md", Hash: "v2", Content: "personal info"},
		},
	}
	parser := &StubParser{Items: []Document{{FilePath: "work/note.md", Hash: "v1", Content: "test"}}}
	embedder := &SpyEmbedder{vector: []float32{0.1, 0.2}}
	formatter := &DefaultFormatter{}
	engine := NewRagEngine(repo, store, parser, tokenizer, embedder, formatter)

	scope := FolderScope{Path: "work/"}
	_, err := engine.Ask(ctx, AskQuery{Question: "test", Scope: scope})
	assert.NoError(t, err)
	assert.Equal(t, scope, store.LastSearchScope)
}

func TestRagEngine_Ask_WithGeneration(t *testing.T) {
	ctx := context.Background()

	store := &SpyVectorStore{
		Documents: []Document{{Content: "найденный контент"}},
	}
	generator := &SpyGenerator{Answer: "ответ от ИИ"}
	contextBuilder := &StubContextBuilder{}

	engine := NewRagEngine(nil, store, nil, &StubTokenizer{}, &SpyEmbedder{}, &DefaultFormatter{})
	engine.SetGenerator(generator, contextBuilder)

	query := AskQuery{
		Question: "как дела?",
		Scope:    AllScope{},
		Generate: true,
	}

	answer, err := engine.Ask(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, "ответ от ИИ", answer)
}
