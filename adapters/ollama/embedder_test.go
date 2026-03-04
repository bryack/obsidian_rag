package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bryack/obsidian_rag/adapters/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	modelName = "argus-ai/pplx-embed-v1-0.6b:fp32"
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
