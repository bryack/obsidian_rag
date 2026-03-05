package markdown

import (
	"fmt"
	"strings"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

type MDParser struct {
	goldmark        goldmark.Markdown
	chunkSize       int
	mergeChunkLimit int
	minChunkSize    int
}

func NewMDParser(chunkSize int, mergeChunkLimit int, minChunkSize int) *MDParser {
	goldmark := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	return &MDParser{
		goldmark:        goldmark,
		chunkSize:       chunkSize,
		mergeChunkLimit: mergeChunkLimit,
		minChunkSize:    minChunkSize,
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
		for child := docNode.FirstChild(); child != nil; child = child.NextSibling() {
			if child.Lines().Len() > 0 {
				start := child.Lines().At(0).Start
				cleanContent = string(source[start:])
				break
			}
		}
	}

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(p.chunkSize),
		textsplitter.WithChunkOverlap(300),
		// textsplitter.WithSeparators([]string{"\n\n", "\n", " ", ""}),
	)

	textChunks, err := splitter.SplitText(cleanContent)
	if err != nil {
		return nil, fmt.Errorf("failed to split text: %w", err)
	}

	mergeChunks := mergeChunks(textChunks, p.mergeChunkLimit)
	var filteredChunks []string

	for _, chunk := range mergeChunks {
		if len(chunk) >= p.minChunkSize {
			filteredChunks = append(filteredChunks, chunk)
		}
	}
	if len(filteredChunks) == 0 {
		filteredChunks = append(filteredChunks, "")
	}

	var docs []domain.Document
	for _, t := range filteredChunks {
		docs = append(docs, domain.Document{
			FilePath: doc.FilePath,
			Hash:     doc.Hash,
			Metadata: meta,
			Content:  t,
		})
	}

	return docs, nil
}

func mergeChunks(raw []string, limit int) []string {
	var merged []string
	var current strings.Builder

	for _, s := range raw {
		if current.Len()+len(s)+2 > limit && current.Len() > 0 {
			merged = append(merged, current.String())
			current.Reset()
		}

		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(s)
	}

	if current.Len() > 0 {
		merged = append(merged, current.String())
	}
	return merged
}
