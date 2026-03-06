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
	"go.abhg.dev/goldmark/wikilink"
)

type MDParser struct {
	goldmark        goldmark.Markdown
	chunkSize       int
	mergeChunkLimit int
	minChunkSize    int
}

func NewMDParser(chunkSize int, mergeChunkLimit int, minChunkSize int) *MDParser {
	goldmark := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}, &wikilink.Extender{}))
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

	rawChunks := p.extractSections(docNode, source)
	merged := mergeChunks(rawChunks, p.mergeChunkLimit)
	filtered := p.filterSmallChunks(merged)

	var docs []domain.Document

	for _, c := range filtered {
		content := p.buildEmbeddingText(doc.FilePath, c.HeaderPath, c.Content)
		chunkMeta := meta
		chunkMeta.Links = uniqueStrings(append(chunkMeta.Links, c.Links...))

		docs = append(docs, domain.Document{
			FilePath:   doc.FilePath,
			Hash:       doc.Hash,
			Content:    content,
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

func (p *MDParser) extractSections(root ast.Node, source []byte) []chunk {
	var chunks []chunk
	var buf strings.Builder
	var headerPath []string
	var currentLinks []string

	flush := func() {
		text := strings.TrimSpace(buf.String())
		if text == "" {
			buf.Reset()
			currentLinks = nil
			return
		}

		chunks = append(chunks, chunk{
			Content:    text,
			HeaderPath: append([]string{}, headerPath...),
			Links:      uniqueStrings(currentLinks),
		})

		buf.Reset()
		currentLinks = nil
	}

	for node := root.FirstChild(); node != nil; node = node.NextSibling() {
		switch n := node.(type) {

		case *ast.Heading:
			if buf.Len() > 0 {
				flush()
			}

			title, links := p.extractNodeText(n, source)
			currentLinks = append(currentLinks, links...)
			headerPath = updateHeaderPath(headerPath, n.Level, title)

			buf.WriteString(title)
			buf.WriteString("\n\n")

		default:
			text, links := p.extractNodeText(n, source)

			if text != "" {
				buf.WriteString(text)
				buf.WriteString("\n\n")
				currentLinks = append(currentLinks, links...)
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
		current.Links = append(current.Links, c.Links...)
		buffer.WriteString(c.Content)
	}

	if buffer.Len() > 0 {
		current.Content = buffer.String()
		merged = append(merged, current)
	}

	return merged
}

func (p *MDParser) extractNodeText(node ast.Node, source []byte) (string, []string) {
	var builder strings.Builder
	linksMap := make(map[string]struct{})

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := n.(*ast.Text); ok {
				builder.Write(t.Value(source))
			}

			if wl, ok := n.(*wikilink.Node); ok {
				linkName := string(wl.Target)
				if linkName != "" {
					linksMap[linkName] = struct{}{}
				}
			}
		}
		return ast.WalkContinue, nil
	})

	var links []string
	for l := range linksMap {
		links = append(links, l)
	}

	return builder.String(), links
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

func uniqueStrings(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	m := make(map[string]struct{})
	var result []string
	for _, s := range input {
		if _, ok := m[s]; !ok {
			m[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}
