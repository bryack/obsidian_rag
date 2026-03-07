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
	mergeChunkLimit = 1500
	minChunkSize    = 50
)

func TestNewMDParser(t *testing.T) {
	t.Run("returns error for invalid parameters", func(t *testing.T) {
		_, err := NewMDParser(-1, mergeChunkLimit, minChunkSize)
		assert.Error(t, err)
	})
}

func TestParse(t *testing.T) {
	t.Run("simple content", func(t *testing.T) {
		testDoc := domain.Document{
			FilePath: "simple_content.md",
			Content:  "Hello World. And some text to be a good chunk to parse",
		}

		parser, err := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		assert.NoError(t, err)
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
			FilePath: "real_content.md",
			Content:  contentWithYAML,
		}
		parser, err := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		assert.NoError(t, err)
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

		parser, err := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		assert.NoError(t, err)
		chunks, err := parser.Parse(testDoc)
		assert.NoError(t, err)

		assert.True(t, len(chunks) > 1, "Expected content to be split into multiple chunks")
		for i, chunk := range chunks {
			assert.Equal(t, "testdata.md", chunk.FilePath, "Chunk %d missing FilePath", i)
			assert.Equal(t, []string{"tdd"}, chunk.Metadata.Tags, "Chunk %d missing Tags", i)
			assert.NotEmpty(t, chunk.Content, "Chunk %d is empty", i)
			assert.LessOrEqual(t, len(chunk.Content), maxChunkSize+2000, "Chunk too large")
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
		parser, err := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		assert.NoError(t, err)

		docs, err := parser.Parse(testDoc)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(docs))
		assert.Equal(t, "", docs[0].Content)
	})

	t.Run("respect min chunk size", func(t *testing.T) {
		testDoc := domain.Document{
			Content: "Hello World",
		}

		parser, err := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
		assert.NoError(t, err)
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

	parser, err := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
	assert.NoError(t, err)
	docs, err := parser.Parse(testDoc)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(docs))
}

func TestMDParser_HeadingSplitting(t *testing.T) {
	content := `---
tags: [test]
---
                                    
## Заголовок второго уровня
Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке. Это текст должен быть в первом чанке.
## Совсем другой текст в заголовке
Это текст должен быть во втором чанке. Это текст должен быть во втором чанке. Это текст должен быть во втором чанке. Это текст должен быть во втором чанке. Это текст должен быть во втором чанке. Это текст должен быть во втором чанке. Это текст должен быть во втором чанке.`

	testDoc := domain.Document{
		Content: content,
	}

	parser, err := NewMDParser(500, 1000, minChunkSize)
	assert.NoError(t, err)
	docs, err := parser.Parse(testDoc)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(docs))
}

func TestMDParser_ListAggregation(t *testing.T) {
	content := `---
tags: [test]
---
                                    
 - это список. каждый пункт должен быть длинным;
 - второй пункт в списке. попробуем сделать так, чтобы пункты отличались друг от друга;
 - третий;

 - чуть подлиннее;
 - в списке должны быть разные пункты;

 - уже шестой пункт. пусть будет такой длины;

 - сделаем пункт eight немного длинее, чтобы точно знать, что целостность списков сохраняется.
 - предложения должны быть довольно большими, чтобы не оставалось сомнений в том, что парсер работает корректно;
 - и самый последний пункт, который докажет на 100%, что мы всё сделали верно.
`
	testDoc := domain.Document{
		Content: content,
	}

	parser, err := NewMDParser(maxChunkSize, mergeChunkLimit, minChunkSize)
	assert.NoError(t, err)
	docs, err := parser.Parse(testDoc)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(docs))
}

func TestMDParser_Wikilinks(t *testing.T) {
	content := `Check this [[My Note]] and this [[Other Note|Alias]] and also [[Note With Fragment#Section]].`
	testDoc := domain.Document{Content: content}

	parser, err := NewMDParser(500, 1000, 5)
	assert.NoError(t, err)
	docs, err := parser.Parse(testDoc)

	assert.NoError(t, err)
	assert.Contains(t, docs[0].Metadata.Links, "My Note")
	assert.Contains(t, docs[0].Metadata.Links, "Other Note")
	assert.Contains(t, docs[0].Metadata.Links, "Note With Fragment")
}
