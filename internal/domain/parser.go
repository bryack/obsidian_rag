package domain

type Parser interface {
	Parse(doc Document) ([]Document, error)
}
