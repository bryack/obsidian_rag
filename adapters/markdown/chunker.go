package markdown

import "strings"

type Chunker struct {
	mergeLimit int
	minSize    int
}

func NewChunker(mergeLimit, minSize int) *Chunker {
	return &Chunker{
		mergeLimit: mergeLimit,
		minSize:    minSize,
	}
}

func (ch *Chunker) Merge(raw []chunk) []chunk {
	merged := make([]chunk, 0, len(raw))
	var buffer strings.Builder
	var current chunk

	for _, c := range raw {
		if buffer.Len()+len(c.Content)+2 > ch.mergeLimit && buffer.Len() > 0 {
			current.Content = buffer.String()
			merged = append(merged, current)
			buffer.Reset()
			current = chunk{}
		}

		if len(c.HeaderPath) > 0 {
			current.HeaderPath = append([]string{}, c.HeaderPath...)
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

func (ch *Chunker) Filter(chunks []chunk) []chunk {
	var result []chunk

	for _, chunk := range chunks {
		if len(chunk.Content) < ch.minSize {
			continue
		}
		result = append(result, chunk)
	}
	return result
}
