package markdown

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeChunks(t *testing.T) {
	t.Run("merges small chunks", func(t *testing.T) {
		chunker := NewChunker(100, 100, 50)
		raw := []chunk{
			{Content: "Part 1", HeaderPath: []string{"H1"}, Links: []string{}},
			{Content: "Part 2", HeaderPath: []string{"H1", "H2"}, Links: []string{"Ссылка"}},
		}

		got := chunker.Merge(raw)
		want := []chunk{{Content: "Part 1\n\nPart 2", HeaderPath: []string{"H1", "H2"}, Links: []string{"Ссылка"}}}
		assert.Equal(t, want, got)
	})

	t.Run("exceeds limit", func(t *testing.T) {
		chunker := NewChunker(100, 100, 20)
		raw := []chunk{
			{
				Content:    "Это текст должен быть в первом чанке. Это текст должен быть в первом чанке.",
				HeaderPath: []string{"Заголовок первого уровня", "Заголовок второго уровня"},
			},
			{
				Content:    "Это текст должен быть во втором чанке. Это текст должен быть во втором чанке.",
				HeaderPath: []string{"Заголовок первого уровня", "Заголовок второго уровня"},
			},
		}

		got := chunker.Merge(raw)
		assert.Equal(t, raw, got)
	})

	t.Run("empty raw", func(t *testing.T) {
		chunker := NewChunker(50, 50, 20)
		raw := []chunk{}

		got := chunker.Merge(raw)
		assert.Equal(t, raw, got)
	})

	t.Run("splits huge section", func(t *testing.T) {
		chunker := NewChunker(1000, 1500, 50)
		hugeContent := fmt.Sprintf("# Заголовок\n %s", strings.Repeat("A", 3000))

		raw := []chunk{{Content: hugeContent, HeaderPath: []string{"Заголовок"}}}
		got := chunker.Merge(raw)

		assert.True(t, len(got) >= 3)
		for _, c := range got {
			assert.LessOrEqual(t, len(c.Content), 1200, "Each chunk should be around chunkSize")
		}
	})
}

func TestFilterSmallChunks(t *testing.T) {
	chunker := NewChunker(1000, 1500, 50)
	chunks := []chunk{
		{
			Content:    "Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке.",
			HeaderPath: []string{"Заголовок первого уровня", "Заголовок второго уровня"},
		},
		{
			Content:    "Слишком маленький чанк",
			HeaderPath: []string{"Заголовок первого уровня", "Заголовок второго уровня"},
		},
	}

	got := chunker.Filter(chunks)
	want := []chunk{
		{
			Content:    "Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке.",
			HeaderPath: []string{"Заголовок первого уровня", "Заголовок второго уровня"},
		},
	}

	assert.Equal(t, want, got)
}
