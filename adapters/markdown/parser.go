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

	rawChunks := p.extractSections(docNode, source)
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
	return append(current, title)
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
