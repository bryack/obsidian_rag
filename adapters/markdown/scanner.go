package markdown

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"go.abhg.dev/goldmark/wikilink"
)

type MDScanner struct {
	levels []int
}

func NewMDScanner() *MDScanner {
	return &MDScanner{levels: []int{}}
}

func (s *MDScanner) Scan(root ast.Node, source []byte) []chunk {
	chunks := make([]chunk, 0, 10)
	var buf strings.Builder
	buf.Grow(len(source))
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

			title, links := extractNodeText(n, source)
			currentLinks = append(currentLinks, links...)
			headerPath, s.levels = updateHeaderPath(headerPath, s.levels, n.Level, title)

			buf.WriteString(title)
			buf.WriteString("\n\n")

		default:
			text, links := extractNodeText(n, source)

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

func extractNodeText(node ast.Node, source []byte) (string, []string) {
	var builder strings.Builder
	builder.Grow(len(source))
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

func updateHeaderPath(current []string, levels []int, level int, title string) ([]string, []int) {
	cutoff := 0
	for i, l := range levels {
		if l >= level {
			break
		}
		cutoff = i + 1
	}

	current = current[:cutoff]
	levels = levels[:cutoff]

	return append(current, title), append(levels, level)
}
