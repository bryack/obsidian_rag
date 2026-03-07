package markdown

import (
	"fmt"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
	"go.abhg.dev/goldmark/wikilink"
)

type MDParser struct {
	goldmark        goldmark.Markdown
	chunker         *Chunker
	chunkSize       int
	mergeChunkLimit int
	minChunkSize    int
}

func NewMDParser(chunkSize, mergeChunkLimit, minChunkSize int) (*MDParser, error) {
	if chunkSize <= 0 || mergeChunkLimit <= 0 || minChunkSize <= 0 {
		return nil, fmt.Errorf("require positive number")
	}

	gm := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}, &wikilink.Extender{}))

	return &MDParser{
		goldmark:  gm,
		chunker:   NewChunker(mergeChunkLimit, minChunkSize),
		chunkSize: chunkSize,
	}, nil
}

type chunk struct {
	Content    string
	HeaderPath []string
	Links      []string
}

func (p *MDParser) Parse(doc domain.Document) ([]domain.Document, error) {
	source := []byte(doc.Content)
	ctx := parser.NewContext()
	reader := text.NewReader(source)

	docNode := p.goldmark.Parser().Parse(reader, parser.WithContext(ctx))

	var meta domain.Metadata
	if d := frontmatter.Get(ctx); d != nil {
		if err := d.Decode(&meta); err != nil {
			return nil, fmt.Errorf("failed to decode frontmatter: %w", err)
		}
	}

	scanner := NewMDScanner()
	rawChunks := scanner.Scan(docNode, source)
	merged := p.chunker.Merge(rawChunks)
	filtered := p.chunker.Filter(merged)

	var docs []domain.Document

	for _, c := range filtered {
		chunkMeta := meta
		chunkMeta.Links = uniqueStrings(append(chunkMeta.Links, c.Links...))

		docs = append(docs, domain.Document{
			FilePath:   doc.FilePath,
			Hash:       doc.Hash,
			Content:    c.Content,
			HeaderPath: c.HeaderPath,
			Metadata:   chunkMeta,
		})
	}
	if len(docs) == 0 {
		docs = append(docs, domain.Document{
			FilePath:   doc.FilePath,
			Hash:       doc.Hash,
			Content:    "",
			HeaderPath: nil,
			Metadata:   meta,
		})
	}

	return docs, nil
}

func uniqueStrings(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	m := make(map[string]struct{})
	result := make([]string, 0, len(input))
	for _, s := range input {
		if _, ok := m[s]; !ok {
			m[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}
