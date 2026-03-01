package markdown

import (
	"fmt"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

type MDParser struct {
	goldmark  goldmark.Markdown
	chunkSize int
}

func NewMDParser(chunkSize int) *MDParser {
	goldmark := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	return &MDParser{
		goldmark:  goldmark,
		chunkSize: chunkSize,
	}
}

func (p *MDParser) Parse(doc domain.Document) ([]domain.Document, error) {
	source := []byte(doc.Content)
	ctx := parser.NewContext()
	reader := text.NewReader([]byte(doc.Content))

	docNode := p.goldmark.Parser().Parse(reader, parser.WithContext(ctx))

	var meta domain.Metadata
	if d := frontmatter.Get(ctx); d != nil {
		if err := d.Decode(&meta); err != nil {
			return nil, fmt.Errorf("failed to decode frontmatter: %w", err)
		}
	}

	cleanContent := ""
	if docNode.HasChildren() {
		start := docNode.FirstChild().Lines().At(0).Start
		cleanContent = string(source[start:])
	}

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(p.chunkSize),
		textsplitter.WithChunkOverlap(0),
	)

	textChunks, err := splitter.SplitText(cleanContent)
	if err != nil {
		return nil, fmt.Errorf("failed to split text: %w", err)
	}

	var docs []domain.Document
	for _, t := range textChunks {
		docs = append(docs, domain.Document{
			FilePath: doc.FilePath,
			Hash:     doc.Hash,
			Metadata: meta,
			Content:  t,
		})
	}

	return docs, nil
}
