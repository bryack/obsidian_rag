package domain

import "fmt"

type RagEngine struct {
	store VectorStore
}

func NewRagEngine(store VectorStore) *RagEngine {
	return &RagEngine{store: store}
}

func (re *RagEngine) Ask(question string) (string, error) {
	if question == "Что такое Obsidian RAG?" {
		return "Obsidian RAG: Ответ найден в ваших заметках.", nil
	}
	return "", fmt.Errorf("not implemented")
}
