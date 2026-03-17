package statsrepo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bryack/obsidian_rag/internal/domain"
)

type FileStatsRepository struct {
	path string
}

func NewFileStatsRepository(vaultPath string) *FileStatsRepository {
	return &FileStatsRepository{
		path: filepath.Join(vaultPath, ".stats", "bm25_stats.json"),
	}
}

func (r FileStatsRepository) Save(stats *domain.BM25Stats) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	err = os.WriteFile(r.path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %q: %w", r.path, err)
	}

	return nil
}

func (r FileStatsRepository) Load() (*domain.BM25Stats, error) {
	data, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("stats not found, run index first")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", r.path, err)
	}
	stats := domain.BM25Stats{}
	err = json.Unmarshal(data, &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return &stats, nil
}
