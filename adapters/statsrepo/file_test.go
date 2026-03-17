package statsrepo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStatsRepository(t *testing.T) {

	t.Run("Save and Load", func(t *testing.T) {
		tempDir := t.TempDir()

		stats := domain.NewBM25Stats(1.5, 0.75)
		stats.DocsNumber = 100
		stats.AverageLength = 50
		stats.DocFrequency = map[string]int{"docker": 15, "go": 10}

		statsRepo := NewFileStatsRepository(tempDir)
		err := statsRepo.Save(stats)
		assert.NoError(t, err)
		want, err := statsRepo.Load()
		assert.NoError(t, err)

		assert.Equal(t, want, stats)
	})

	t.Run("Save creates a directory", func(t *testing.T) {
		tempDir := t.TempDir()
		statsRepo := NewFileStatsRepository(tempDir)
		err := statsRepo.Save(domain.NewBM25Stats(1.5, 0.75))
		assert.NoError(t, err)

		entries, err := os.ReadDir(filepath.Join(tempDir, ".stats"))
		assert.NoError(t, err)
		assert.Len(t, entries, 1)
	})

	t.Run("Load returns error if file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		statsRepo := NewFileStatsRepository(tempDir)
		err := statsRepo.Save(domain.NewBM25Stats(1.5, 0.75))
		err = os.Remove(filepath.Join(tempDir, ".stats", "bm25_stats.json"))
		require.NoError(t, err)
		_, err = statsRepo.Load()
		assert.Error(t, err)
	})

	t.Run("resave rewrite file", func(t *testing.T) {
		tempDir := t.TempDir()

		stats := domain.NewBM25Stats(1.5, 0.75)
		stats.DocsNumber = 100
		stats.AverageLength = 50
		stats.DocFrequency = map[string]int{"docker": 15, "go": 10}

		statsRepo := NewFileStatsRepository(tempDir)
		err := statsRepo.Save(stats)
		assert.NoError(t, err)

		stats.DocsNumber = 200
		err = statsRepo.Save(stats)
		assert.NoError(t, err)

		want, err := statsRepo.Load()
		assert.NoError(t, err)

		assert.Equal(t, 200, want.DocsNumber)
	})
}
