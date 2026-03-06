package markdown

import (
	"fmt"
	"strings"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
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

type chunk struct {
	Content    string
	HeaderPath []string
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

	rawChunks := p.extractSections(docNode, source)
	merged := mergeChunks(rawChunks, p.mergeChunkLimit)
	filtered := p.filterSmallChunks(merged)

	var docs []domain.Document

	for _, c := range filtered {
		content := p.buildEmbeddingText(doc.FilePath, c.HeaderPath, c.Content)

		docs = append(docs, domain.Document{
			FilePath:   doc.FilePath,
			Hash:       doc.Hash,
			Content:    content,
			HeaderPath: c.HeaderPath,
			Metadata:   meta,
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

func (p *MDParser) extractSections(root ast.Node, source []byte) []chunk {
	var chunks []chunk
	var buf strings.Builder
	var headerPath []string

	flush := func() {
		text := strings.TrimSpace(buf.String())
		if text == "" {
			buf.Reset()
			return
		}

		chunks = append(chunks, chunk{
			Content:    text,
			HeaderPath: append([]string{}, headerPath...),
		})

		buf.Reset()
	}

	for node := root.FirstChild(); node != nil; node = node.NextSibling() {

		switch n := node.(type) {

		case *ast.Heading:

			if buf.Len() > 0 {
				flush()
			}

			title := p.extractNodeText(n, source)

			headerPath = updateHeaderPath(headerPath, n.Level, title)

			buf.WriteString(title)
			buf.WriteString("\n\n")

		default:

			text := p.extractNodeText(n, source)

			if text != "" {
				buf.WriteString(text)
				buf.WriteString("\n\n")
			}
		}
	}

	flush()

	return chunks
}

func mergeChunks(raw []chunk, limit int) []chunk {
	var merged []chunk
	var buffer strings.Builder
	var current chunk

	for _, c := range raw {
		if buffer.Len()+len(c.Content)+2 > limit && buffer.Len() > 0 {
			current.Content = buffer.String()
			merged = append(merged, current)
			buffer.Reset()
			current = chunk{}
		}

		if buffer.Len() == 0 {
			current.HeaderPath = append(current.HeaderPath, c.HeaderPath...)
		}

		if buffer.Len() > 0 {
			buffer.WriteString("\n\n")
		}

		buffer.WriteString(c.Content)
	}

	if buffer.Len() > 0 {
		current.Content = buffer.String()
		merged = append(merged, current)
	}

	return merged
}

func (p *MDParser) extractNodeText(node ast.Node, source []byte) string {
	var builder strings.Builder

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := n.(*ast.Text); ok {
				builder.Write(t.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})

	return builder.String()
}

func updateHeaderPath(current []string, level int, title string) []string {
	if level <= len(current) {
		current = current[:level-1]
	}

	for len(current) < level-1 {
		current = append(current, "")
	}
	return append(current, title)
}

func (p *MDParser) filterSmallChunks(chunks []chunk) []chunk {
	var result []chunk

	for _, chunk := range chunks {
		if len(chunk.Content) < p.minChunkSize {
			continue
		}
		result = append(result, chunk)
	}
	return result
}

func (p *MDParser) buildEmbeddingText(file string, headers []string, content string) string {
	var builder strings.Builder

	builder.WriteString("File: ")
	builder.WriteString(file)
	builder.WriteString("\n")

	if len(headers) > 0 {
		builder.WriteString("Section: ")
		builder.WriteString(strings.Join(headers, " / "))
		builder.WriteString("\n\n")
	}

	builder.WriteString(content)
	return builder.String()
}
