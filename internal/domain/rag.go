package domain

import "fmt"

type RagEngine struct{}

func NewRagEngine() *RagEngine {
	return &RagEngine{}
}

func (re *RagEngine) Ask(question string) (string, error) {
	if question == "Что такое Obsidian RAG?" {
		return "Obsidian RAG: Ответ найден в ваших заметках.", nil
	}
	return "", fmt.Errorf("not implemented")
}
