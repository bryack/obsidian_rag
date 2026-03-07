package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFormatter_Format(t *testing.T) {
	formatter := &DefaultFormatter{}

	t.Run("with headers", func(t *testing.T) {
		doc := Document{
			FilePath:   "test.md",
			HeaderPath: []string{"H1", "H2"},
			Content:    "Hello world",
		}

		got := formatter.Format(doc)
		want := "File: test.md\nSection: H1 / H2\n\nHello world"

		assert.Equal(t, want, got)
	})

	t.Run("without headers", func(t *testing.T) {
		doc := Document{
			FilePath:   "without_headers.md",
			HeaderPath: []string{},
			Content:    "Hello world",
		}

		got := formatter.Format(doc)
		want := "File: without_headers.md\nHello world"

		assert.Equal(t, want, got)
	})
}
