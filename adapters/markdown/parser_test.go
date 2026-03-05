package markdown

import (
	"os"
	"testing"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	maxChunkSize    = 1000
	mergeChunkLimit = 2000
	minChunkSize    = 50
)

func TestParse(t *testing.T) {
	t.Run("simple content", func(t *testing.T) {
		testDoc := domain.Document{
			Content: "Hello World. And some text to be a good chunk to parse",
		}

		parser := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		docs, err := parser.Parse(testDoc)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(docs))
		assert.Equal(t, "Hello World. And some text to be a good chunk to parse", docs[0].Content)
	})

	t.Run("with frontmatter", func(t *testing.T) {
		contentWithYAML := `---
tags:
  - obsidian
  - rag
project: [obsidian-rag]
---
Real content here. And some text to be a good chunk to parse`

		testDoc := domain.Document{
			Content: contentWithYAML,
		}
		parser := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		docs, err := parser.Parse(testDoc)
		assert.NoError(t, err)

		assert.Equal(t, []string{"obsidian", "rag"}, docs[0].Metadata.Tags)
		assert.Equal(t, []string{"obsidian-rag"}, docs[0].Metadata.Project)
		assert.Equal(t, "Real content here. And some text to be a good chunk to parse", docs[0].Content)
	})

	t.Run("splitting", func(t *testing.T) {
		testFile := "testdata.md"
		testData, err := os.ReadFile(testFile)
		require.NoError(t, err)

		testDoc := domain.Document{
			FilePath: testFile,
			Content:  string(testData),
		}

		parser := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		chunks, err := parser.Parse(testDoc)
		assert.NoError(t, err)

		assert.True(t, len(chunks) > 1, "Expected content to be split into multiple chunks")
		for i, chunk := range chunks {
			assert.Equal(t, "testdata.md", chunk.FilePath, "Chunk %d missing FilePath", i)
			assert.Equal(t, []string{"tdd"}, chunk.Metadata.Tags, "Chunk %d missing Tags", i)
			assert.NotEmpty(t, chunk.Content, "Chunk %d is empty", i)
			assert.LessOrEqual(t, len(chunk.Content), maxChunkSize+1500, "Chunk too large")
		}
	})
	t.Run("empty content after yaml", func(t *testing.T) {
		content := `---
tags: [test]
---
# `

		testDoc := domain.Document{
			Content: content,
		}
		parser := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)

		docs, err := parser.Parse(testDoc)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(docs))
		assert.Equal(t, "", docs[0].Content)
	})

	t.Run("respect min chunk size", func(t *testing.T) {
		testDoc := domain.Document{
			Content: "Hello World",
		}

		parser := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		docs, err := parser.Parse(testDoc)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(docs))
		assert.Equal(t, "", docs[0].Content)
	})
}

func TestMDParser_Aggregation(t *testing.T) {
	content := `---
tags: [test]
---
                                    
Это текст, который должен быть в одном чанке, а не в нескольких.
                                                                
                                      
[00:35:56] Это текст, который должен быть в одном чанке, а не в нескольких.
                    
                                             
[00:35:56] Это текст, который должен быть в одном чанке, а не в нескольких.
                                           
                                                          
## Тут заголовок, который относится к этому же чанку
[00:35:56] Это текст, который должен быть в одном чанке, а не в нескольких.
                           

                                                  
### Тут небольшой заголовок
[00:35:56] Это текст, который должен быть в одном чанке, а не в нескольких.`

	testDoc := domain.Document{
		Content: content,
	}

	parser := NewMDParser(500, mergeChunkLimit, minChunkSize)
	docs, err := parser.Parse(testDoc)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(docs))
}
