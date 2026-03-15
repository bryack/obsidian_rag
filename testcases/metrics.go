package testcases

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/bryack/obsidian_rag/internal/domain"
)

type EvaluationResult struct {
	TestCase      TestCase
	Docs          []domain.Document
	RelevantFound int
	PrecisionK    float64
	RecallK       float64
	MRR           float64
	Passed        bool
	FailureReason string
}

func Evaluate(tc TestCase, docs []domain.Document, paramK int) (*EvaluationResult, error) {
	result := &EvaluationResult{
		TestCase: tc,
		Docs:     docs,
	}
	if len(docs) > paramK {
		docs = docs[:paramK]
	}

	rank := make([]int, 0, len(docs))
	for i := range docs {
		isRelevant := false
		for _, chunk := range tc.RelevantChunks {
			if ok := chunk.Match(docs[i]); !ok {
				continue
			}
			isRelevant = true
			result.RelevantFound++
			break
		}
		if isRelevant {
			rank = append(rank, i+1)
		}
	}

	result.PrecisionK = float64(result.RelevantFound) / float64(paramK)

	result.RecallK = float64(result.RelevantFound) / float64(len(tc.RelevantChunks))
	if len(rank) == 0 {
		result.MRR = 0
	} else {
		result.MRR = 1. / float64(rank[0])
	}

	result.Passed = true
	if result.PrecisionK < tc.MinPrecisionAt5 {
		result.Passed = false
		result.FailureReason = fmt.Sprintf("precision@%d: %.2f (%d relevant), required: %.2f (%d relevant)",
			paramK,
			result.PrecisionK,
			result.RelevantFound,
			tc.MinPrecisionAt5,
			int(tc.MinPrecisionAt5*float64(paramK)),
		)
	}
	return result, nil
}

func (c *ChunkID) Match(doc domain.Document) bool {
	if normalizeFilePath(doc.FilePath) != normalizeFilePath(c.FilePath) {
		return false
	}

	if len(c.HeaderPath) != 0 {
		if !compareSlices(doc.HeaderPath, c.HeaderPath) {
			return false
		}
	}
	return true
}

func normalizeFilePath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func compareSlices(A, B []string) bool {
	for _, a := range A {
		if slices.Contains(B, a) {
			return true
		}
	}
	return false
}
