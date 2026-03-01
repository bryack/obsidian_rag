package main

import (
	"fmt"
	"os"

	"github.com/bryack/obsidian_rag/adapters/filerepo"
	"github.com/bryack/obsidian_rag/internal/domain"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: obsidian-rag <command> <vault_path> [question]")
		return
	}

	command := os.Args[1]
	vaultPath := os.Args[2]

	store := &domain.SpyVectorStore{}
	repo := filerepo.NewRepository(os.DirFS(vaultPath))
	engine := domain.NewRagEngine(repo, store)

	switch command {
	case "index":
		err := engine.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Index error: %v\n", err)
			os.Exit(1)
		}
	case "ask":
		if len(os.Args) < 4 {
			fmt.Println("Usage: obsidian-rag ask <vault_path> <question>")
			return
		}
		err := engine.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Index error: %v\n", err)
			os.Exit(1)
		}
		answer, err := engine.Ask(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(answer)
	default:
		fmt.Printf("Unknown command: %q\n", command)

	}
}
