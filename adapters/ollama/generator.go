package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaGenerator struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaGenerator(baseURL, model string) *OllamaGenerator {
	return &OllamaGenerator{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

type generateRequest struct {
	Model     string                 `json:"model"`
	Prompt    string                 `json:"prompt"`
	Stream    bool                   `json:"stream"`
	Options   map[string]interface{} `json:"options"`
	KeepAlive string                 `json:"keep_alive"`
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (og *OllamaGenerator) Generate(ctx context.Context, question, contextText string) (string, error) {
	prompt := fmt.Sprintf(`Ты — аналитик заметок Obsidian. Твоя задача — найти ответ в КОНТЕКСТЕ.

   ПРАВИЛА:
   - Отвечай на русском языке.
   - Отвечай развёрнуто.
   - Будь гибким в названиях: название может быть с заглавными или строчными буквами - это не важно.
   - Если в тексте есть хоть какая-то информация, используй её.
   - Если информации совсем нет, тогда ответь "В моих заметках нет информации".

КОНТЕКСТ:
%s

ВОПРОС: %s
`, contextText, question)

	reqBody := generateRequest{
		Model:  og.model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.5,
			"num_ctx":     16384,
		},
		KeepAlive: "10m",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", og.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := og.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned error status: %d", resp.StatusCode)
	}

	var genResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return genResp.Response, nil
}
