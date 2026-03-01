package filerepo

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/bryack/obsidian_rag/internal/domain"
)

type Repository struct {
	fileSystem fs.FS
}

func NewRepository(fileSystem fs.FS) *Repository {
	return &Repository{fileSystem: fileSystem}
}

func (r *Repository) GetNotes() ([]domain.Document, error) {
	var chunks []domain.Document

	err := fs.WalkDir(r.fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			data, err := fs.ReadFile(r.fileSystem, path)
			if err != nil {
				return err
			}
			h := sha256.New()
			h.Write(data)
			hashSum := fmt.Sprintf("%x", h.Sum(nil))
			chunks = append(chunks, domain.Document{
				FilePath: path,
				Hash:     hashSum,
				Content:  string(data),
			})
		}
		return nil
	})

	return chunks, err
}
