package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateHeaderPath(t *testing.T) {
	current := []string{"H1"}
	level := 3
	title := "H3"

	got := updateHeaderPath(current, level, title)
	want := []string{"H1", "H3"}

	assert.Equal(t, want, got)
}

func TestMergeChunks(t *testing.T) {
	t.Run("merges small chunks", func(t *testing.T) {
		raw := []chunk{
			{Content: "Part 1", HeaderPath: []string{"H1"}, Links: []string{}},
			{Content: "Part 2", HeaderPath: []string{"H1", "H2"}, Links: []string{"Ссылка"}},
		}
		limit := 100

		got := mergeChunks(raw, limit)
		want := []chunk{{Content: "Part 1\n\nPart 2", HeaderPath: []string{"H1", "H2"}, Links: []string{"Ссылка"}}}
		assert.Equal(t, want, got)
	})

	t.Run("exceeds limit", func(t *testing.T) {
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
		limit := 50

		got := mergeChunks(raw, limit)
		assert.Equal(t, raw, got)
	})

	t.Run("empty raw", func(t *testing.T) {
		raw := []chunk{}
		limit := 50

		got := mergeChunks(raw, limit)
		assert.Equal(t, raw, got)
	})
}

func TestFilterSmallChunks(t *testing.T) {
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

	parser := NewMDParser(100, 1500, 50)
	got := parser.filterSmallChunks(chunks)
	want := []chunk{
		{
			Content:    "Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке.",
			HeaderPath: []string{"Заголовок первого уровня", "Заголовок второго уровня"},
		},
	}

	assert.Equal(t, want, got)
}
