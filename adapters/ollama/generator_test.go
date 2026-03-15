package ollama

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaGenerator(t *testing.T) {
	if testing.Short() {
		t.Skip("test requires running Ollama")
	}

	gen := NewOllamaGenerator("http://localhost:11434", "qwen3.5:9b")

	answer, err := gen.Generate(context.Background(), "Какой код от сейфа?", "Секретный код от сейфа в офисе: 998877. Ключ лежит под ковриком.")
	assert.NoError(t, err)
	assert.Contains(t, answer, "998877")
	t.Logf("Ollama answers: %s", answer)
}
