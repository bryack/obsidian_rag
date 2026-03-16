package testcases

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bryack/obsidian_rag/adapters/filerepo"
	"github.com/bryack/obsidian_rag/adapters/markdown"
	"github.com/bryack/obsidian_rag/adapters/ollama"
	"github.com/bryack/obsidian_rag/adapters/qdrant"
	"github.com/bryack/obsidian_rag/adapters/tokenizer"
	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	chunkSize       = 1000
	mergeChunkLimit = 1500
	minChunkSize    = 200
	embedModelName  = "bge-m3:latest"
	qdrantAddr      = "localhost:6334"
	ollamaURL       = "http://localhost:11434/api/embed"
	vaultPath       = "/home/bryack/Documents/Obsidian/bryack"
	GroundTruthPath = "./fixtures/ground_truth.yaml"
)

func TestRetrievalQualityEvaluation(t *testing.T) {
	if testing.Short() {
		t.Skip("Acceptance test requires running Qdrant and Ollama")
	}

	store, err := qdrant.NewQdrantStore(qdrantAddr)
	require.NoError(t, err)

	repo := filerepo.NewRepository(os.DirFS(vaultPath))
	parser, err := markdown.NewMDParser(chunkSize, mergeChunkLimit, minChunkSize)
	require.NoError(t, err)

	embedder := ollama.NewOllamaEmbedder(embedModelName, ollamaURL)
	tokenizer := tokenizer.NewTokenizer()
	formatter := &domain.DefaultFormatter{}

	engine := domain.NewRagEngine(repo, store, parser, tokenizer, embedder, formatter)

	gt, err := LoadGroundTruth(GroundTruthPath)
	require.NoError(t, err)
	K := 5

	var (
		sumPrecision float64
		passedCount  int
		totalTests   = len(gt.TestCases)
	)

	for _, tc := range gt.TestCases {
		validateGroundTruth(t, tc.RelevantChunks)

		t.Run(tc.Name, func(t *testing.T) {
			docs, err := engine.SearchChunks(context.Background(), domain.AskQuery{
				Question: tc.Query,
				Generate: false,
			})
			require.NoError(t, err)

			result, err := Evaluate(tc, docs, K)
			require.NoError(t, err)
			sumPrecision += result.PrecisionK

			t.Logf("Query: %q", tc.Query)
			t.Logf("Description: %q", tc.Description)

			firstRelevantRank := 0
			if result.MRR > 0 {
				firstRelevantRank = int(1. / result.MRR)
			}

			t.Logf("--- Metrics ---\nRelevant found: %d / %d\nPrecision@5: %.2f\nRecall@5: %.2f (%d / %d)\nMRR: %.2f (first relevant at rank %d)",
				result.RelevantFound, K, result.PrecisionK, result.RecallK, result.RelevantFound, len(tc.RelevantChunks), result.MRR, firstRelevantRank)
			logResults(t, docs, tc)
			t.Logf("--- Validation ---\nExpected: %.2f (%d relevant)\nActual: %.2f (%d relevant)",
				tc.MinPrecisionAt5, int(tc.MinPrecisionAt5*float64(K)), result.PrecisionK, result.RelevantFound)
			if result.Passed {
				passedCount++
				t.Log("Status: PASSED")
			} else {
				t.Log("Status: FAILED")
				t.Logf("Failure: %s", result.FailureReason)
			}
			assert.True(t, result.Passed)
		})
	}

	t.Run("Summary", func(t *testing.T) {
		averagePrecision := sumPrecision / float64(totalTests)
		t.Log("--- Summary ---")
		t.Logf("Total tests: %d", totalTests)
		t.Logf("Passed: %d", passedCount)
		t.Logf("Failed: %d", totalTests-passedCount)
		t.Logf("Average Precision@%d: %.2f", K, averagePrecision)
		t.Logf("Required minimum: %.2f", gt.MinAveragePrecision)

		assert.GreaterOrEqual(t, averagePrecision, gt.MinAveragePrecision, "Average precision below threshold")
	})

}

func validateGroundTruth(t *testing.T, relevantChunks []ChunkID) {
	for i := range relevantChunks {
		fullPath := filepath.Join(vaultPath, relevantChunks[i].FilePath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Logf("WARNING: ground truth outdated - file not found: %s", relevantChunks[i].FilePath)
		}
	}
}

func logResults(t *testing.T, docs []domain.Document, tc TestCase) {
	t.Log("--- Results ---\n")
	for i, doc := range docs {
		marker := "[ ]"
		for _, chunk := range tc.RelevantChunks {
			if chunk.Match(doc) {
				marker = "[✓]"
				break
			}
		}

		for _, pattern := range tc.ExcludedPatterns {
			if strings.Contains(doc.FilePath, pattern) {
				marker = "[✗]"
				break
			}
		}

		t.Logf("%s [%d] Score: %.4f | %s\n", marker, i+1, doc.Score, doc.FilePath)
		if len(doc.HeaderPath) > 0 {
			t.Logf("    Section: %v", doc.HeaderPath)
		}
	}
}
