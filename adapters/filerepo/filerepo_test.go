package filerepo

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	mockFS := fstest.MapFS{
		"hello.md": {Data: []byte("world")},
	}
	files, err := mockFS.ReadDir(".")
	assert.NoError(t, err)

	data, err := mockFS.ReadFile("hello.md")

	repo := NewRepository(mockFS)

	chunks, err := repo.GetNotes()
	assert.NoError(t, err)
	assert.Equal(t, len(files), len(chunks))
	assert.Equal(t, string(data), chunks[0].Content)
	assert.NotEmpty(t, chunks[0].Hash)
}
