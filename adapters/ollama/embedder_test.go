package ollama

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"testing"

	"github.com/bryack/obsidian_rag/adapters/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	modelName = "bge-m3:latest"
	baseURL   = "http://localhost:11434"
)

func TestOllama_HealthCheck(t *testing.T) {

	t.Run("ollama launch", func(t *testing.T) {
		response, err := http.Get(baseURL + "/api/tags")
		require.NoError(t, err, "Ollama service is not running at %s", baseURL)
		defer response.Body.Close()
	})

	t.Run("model check", func(t *testing.T) {
		response, err := http.Get(baseURL + "/api/tags")
		require.NoError(t, err, "Ollama service is not running at %s", baseURL)
		defer response.Body.Close()
		var tags struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}

		err = json.NewDecoder(response.Body).Decode(&tags)
		require.NoError(t, err, "Failed to decode response body")

		found := false
		for _, m := range tags.Models {
			if m.Name == modelName {
				found = true
				break
			}
		}
		require.True(t, found, "Model %s not found in Ollama. Run 'ollama pull %s'", modelName, modelName)
	})

	t.Run("VRAM load", func(t *testing.T) {
		warmupEmbedder := NewOllamaEmbedder(modelName, baseURL+"/api/embed")
		_, err := warmupEmbedder.EmbedQuery(context.Background(), "warmup")
		require.NoError(t, err, "Failed to warmup model for VRAM check")

		response, err := http.Get(baseURL + "/api/ps")
		require.NoError(t, err)
		defer response.Body.Close()

		var ps struct {
			Models []struct {
				Name     string `json:"name"`
				SizeVRAM int64  `json:"size_vram"`
			} `json:"models"`
		}

		err = json.NewDecoder(response.Body).Decode(&ps)
		require.NoError(t, err, "Failed to decode response body")

		modelInMemory := false
		for _, m := range ps.Models {
			if m.Name == modelName {
				modelInMemory = true
				assert.Greater(t, m.SizeVRAM, int64(0), "Model is loaded but NOT using VRAM (GPU drivers issue?)")
				break
			}
		}
		assert.True(t, modelInMemory, "Model %s should be in memory after the warmup call", modelName)
	})
}

func TestOllamaEmbedder(t *testing.T) {
	embedder := NewOllamaEmbedder(modelName, baseURL+"/api/embed")

	vector, err := embedder.EmbedQuery(context.Background(), "Привет!")
	assert.NoError(t, err)
	assert.Equal(t, qdrant.VectorSize, len(vector))
}

func TestOllamaEmbedder_Batch(t *testing.T) {
	chunksContent := []string{"привет", "пока"}
	embedder := NewOllamaEmbedder(modelName, baseURL+"/api/embed")

	vectors, err := embedder.EmbedDocuments(context.Background(), chunksContent)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(vectors))

	for _, vector := range vectors {
		assert.Equal(t, 1024, len(vector))
	}
}

func BenchmarkOllamaEmbedder_Batch(b *testing.B) {
	embedder := NewOllamaEmbedder(modelName, baseURL+"/api/embed")
	texts := make([]string, 32)
	for i := range texts {
		texts[i] = "Это тестовая строка для проверки скорости генерации вектора."
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := embedder.EmbedDocuments(context.Background(), texts)
		require.NoError(b, err)
	}
}

func TestOllamaEmbedder_SemanticSimilarity(t *testing.T) {
	const (
		modelName = "bge-m3:latest"
		baseURL   = "http://localhost:11434/api/embed"
	)

	embedder := NewOllamaEmbedder(modelName, baseURL)
	ctx := context.Background()

	t.Run("validates semantic distance in Russian", func(t *testing.T) {
		queries := []string{
			"Как настроить Docker контейнер?",
			"Как сконфигурировать Docker контейнер?", // Парафраз (высокое сходство)
			"Настройка Kubernetes кластера",          // Связанная тема (среднее сходство)
			"Рецепт приготовления пиццы",             // Совершенно другая тема (низкое сходство)
		}

		embeddings, err := embedder.EmbedDocuments(ctx, queries)
		require.NoError(t, err, "Ollama must be running and model must be pulled")
		require.Equal(t, 4, len(embeddings), "Should return exactly 4 embeddings")

		cosineSimilarity := func(a, b []float32) float64 {
			if len(a) != len(b) {
				return 0.0
			}
			var dotProduct, normA, normB float64
			for i := range a {
				valA := float64(a[i])
				valB := float64(b[i])
				dotProduct += valA * valB
				normA += valA * valA
				normB += valB * valB
			}

			if normA == 0 || normB == 0 {
				return 0.0
			}

			return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
		}

		similarityParaphrase := cosineSimilarity(embeddings[0], embeddings[1])
		similarityRelated := cosineSimilarity(embeddings[0], embeddings[2])
		similarityIrrelevant := cosineSimilarity(embeddings[0], embeddings[3])

		t.Logf("--- Semantic Analysis Results ---")
		t.Logf("Query A: %s", queries[0])
		t.Logf("Query B (Paraphrase): %.4f similarity", similarityParaphrase)
		t.Logf("Query C (Related):    %.4f similarity", similarityRelated)
		t.Logf("Query D (Irrelevant): %.4f similarity", similarityIrrelevant)

		assert.Greater(t, similarityParaphrase, 0.8,
			"Paraphrase similarity (%.4f) should be high (> 0.8)", similarityParaphrase)

		assert.Less(t, similarityIrrelevant, 0.4,
			"Irrelevant similarity (%.4f) should be low (< 0.4)", similarityIrrelevant)

		assert.Greater(t, similarityParaphrase, similarityRelated,
			"Paraphrase should be closer than just related topic")

		assert.Greater(t, similarityRelated, similarityIrrelevant,
			"Related topic should be closer than completely irrelevant text")
	})
}
