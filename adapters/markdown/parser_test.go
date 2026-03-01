package markdown

import (
	"testing"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestParse_SimpleContent(t *testing.T) {
	testDoc := domain.Document{
		Content: "Hello World",
	}

	parser := MDParser{}
	docs, err := parser.Parse(testDoc)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(docs))
	assert.Equal(t, "Hello World", docs[0].Content)
}
