package markdown

import (
	"fmt"
	"strings"

	"github.com/bryack/obsidian_rag/internal/domain"
)

type DefaultContextBuilder struct {
	maxChars int
}

func NewDefaultContextBuilder(maxChars int) *DefaultContextBuilder {
	return &DefaultContextBuilder{maxChars: maxChars}
}

func (d *DefaultContextBuilder) BuildContext(chunks []domain.Document) string {
	var builder strings.Builder
	builder.Grow(d.maxChars)
	currentLen := 0

	for i, chunk := range chunks {
		sourceHeader := fmt.Sprintf("[Источник %d: %s]\n", i+1, chunk.FilePath)
		sourceHeaderLen := len(sourceHeader)

		if currentLen+sourceHeaderLen > d.maxChars {
			break
		}

		builder.WriteString(sourceHeader)
		currentLen += sourceHeaderLen

		remainingSpace := d.maxChars - currentLen

		contentLen := len(chunk.Content) + 2
		if contentLen > remainingSpace {
			endPos := remainingSpace
			if endPos > len(chunk.Content) {
				endPos = len(chunk.Content)
			}
			if lastSpace := strings.LastIndex(chunk.Content[:endPos], " "); lastSpace > 0 {
				endPos = lastSpace
			}
			builder.WriteString(chunk.Content[:endPos])
			builder.WriteString("\n[... ОБРЕЗАНО ...]\n\n")
			break
		}

		builder.WriteString(chunk.Content)
		builder.WriteString("\n\n")
		currentLen += contentLen
	}

	return builder.String()
}
