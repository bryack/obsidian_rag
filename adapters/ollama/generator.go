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
}

func NewOllamaGenerator(baseURL, model string) *OllamaGenerator {
	return &OllamaGenerator{baseURL: baseURL, model: model}
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
	prompt := fmt.Sprintf(`Ты — ассистент по базе знаний Obsidian. Твоя задача — отвечать на вопросы пользователя, используя ТОЛЬКО предоставленный ниже контекст. 

ИНСТРУКЦИИ:
1. Отвечай на русском языке.
2. Если в контексте нет ответа на вопрос, честно скажи: "В моих заметках нет информации по этому вопросу".
3. Не используй свои внешние знания, только то, что написано в блоке КОНТЕКСТ.
4. Ответ должен быть лаконичным и структурированным.

КОНТЕКСТ:
%s

ВОПРОС: %s

ОТВЕТ:`, contextText, question)

	reqBody := generateRequest{
		Model:  og.model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.1,
			"num_ctx":     8192,
		},
		KeepAlive: "5m",
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

	resp, err := http.DefaultClient.Do(req)
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
