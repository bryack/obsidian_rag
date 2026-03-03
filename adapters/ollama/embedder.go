package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaEmbedder struct {
	ModelName string
	BaseURL   string
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func NewOllamaEmbedder(modelName, baseURL string) *OllamaEmbedder {
	return &OllamaEmbedder{
		ModelName: modelName,
		BaseURL:   baseURL,
	}
}

func (o *OllamaEmbedder) Embed(text string) ([]float32, error) {
	requestBody, err := json.Marshal(embedRequest{
		Model: o.ModelName,
		Input: text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json request: %w", err)
	}

	response, err := http.Post(o.BaseURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to post response on URL %q: %w", o.BaseURL, err)
	}
	defer response.Body.Close()

	var res embedResponse
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return res.Embeddings[0], nil
}
