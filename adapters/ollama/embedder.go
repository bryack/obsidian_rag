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
	Model string   `json:"model"`
	Input []string `json:"input"`
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

func (o *OllamaEmbedder) EmbedQuery(text string) ([]float32, error) {
	res, err := o.send([]string{text})
	if err != nil {
		return nil, fmt.Errorf("failed to send text to Ollama: %w", err)
	}

	if len(res) < 1 {
		return nil, fmt.Errorf("expected at least 1 embedding, but got 0")
	}
	return res[0], err
}

func (o *OllamaEmbedder) EmbedDocuments(texts []string) ([][]float32, error) {
	res, err := o.send(texts)
	if err != nil {
		return nil, fmt.Errorf("failed to send texts to Ollama: %w", err)
	}

	if len(res) != len(texts) {
		return nil, fmt.Errorf("batch size mismatch: sent %d, got %d", len(texts), len(res))
	}
	return res, nil
}

func (o *OllamaEmbedder) send(texts []string) ([][]float32, error) {
	requestBody, err := json.Marshal(embedRequest{
		Model: o.ModelName,
		Input: texts,
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

	return res.Embeddings, nil
}
