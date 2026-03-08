package markdown

import (
	"testing"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestDefaultContextBuilder(t *testing.T) {
	builder := NewDefaultContextBuilder(300)

	chunks := []domain.Document{
		{
			FilePath: "notes/work.md",
			Content:  "Первая важная заметка.",
		},
		{
			FilePath: "notes/private.md",
			Content:  "Вторая заметка с очень длинным текстом, который должен быть обрезан лимитом, чтобы не переполнять контекст нейросети.",
		},
	}

	result := builder.BuildContext(chunks)
	assert.Contains(t, result, "[Источник 1: notes/work.md]")
	assert.Contains(t, result, "Первая важная заметка.")
	assert.Contains(t, result, "[Источник 2: notes/private.md]")
}
