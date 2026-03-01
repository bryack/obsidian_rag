package markdown

import (
	"github.com/bryack/obsidian_rag/internal/domain"
)

type MDParser struct{}

func (p *MDParser) Parse(doc domain.Document) ([]domain.Document, error) {
	docs := []domain.Document{}
	docs = append(docs, doc)
	return docs, nil
}
