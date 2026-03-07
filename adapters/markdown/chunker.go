package markdown

import (
	"strings"

	"github.com/tmc/langchaingo/textsplitter"
)

const (
	chunkSizeToleranceFactor = 1.5
	chunkOverlapFactor       = 10
)

type Chunker struct {
	chunkSize  int
	mergeLimit int
	minSize    int
	splitter   textsplitter.TextSplitter
}

func NewChunker(chunkSize, mergeLimit, minSize int) *Chunker {
	return &Chunker{
		chunkSize:  chunkSize,
		mergeLimit: mergeLimit,
		minSize:    minSize,
		splitter: textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(chunkSize),
			textsplitter.WithChunkOverlap(chunkSize/chunkOverlapFactor),
		),
	}
}

func (ch *Chunker) Merge(raw []chunk) []chunk {
	merged := make([]chunk, 0, len(raw))
	var buffer strings.Builder
	var current chunk

	addChunk := func(c chunk, content string) {
		if len(content) <= int(float64(ch.chunkSize)*chunkSizeToleranceFactor) {
			c.Content = content
			merged = append(merged, c)
			return
		}

		subText, err := ch.splitter.SplitText(content)
		if err != nil {
			c.Content = content
			merged = append(merged, c)
			return
		}

		for _, sub := range subText {
			newChunk := c
			newChunk.Content = sub
			merged = append(merged, newChunk)
		}

	}

	for _, c := range raw {
		if buffer.Len()+len(c.Content)+2 > ch.mergeLimit && buffer.Len() > 0 {
			addChunk(current, buffer.String())
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
		addChunk(current, buffer.String())
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
