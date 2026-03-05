package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bryack/obsidian_rag/adapters/filerepo"
	"github.com/bryack/obsidian_rag/adapters/markdown"
	"github.com/bryack/obsidian_rag/adapters/ollama"
	"github.com/bryack/obsidian_rag/adapters/qdrant"
	"github.com/bryack/obsidian_rag/internal/domain"
)

const (
	chunkSize       = 1000
	mergeChunkLimit = 2000
	minChunkSize    = 100
	embedModelName  = "argus-ai/pplx-embed-v1-0.6b:fp32"
)

var (
	qdrantAddr = flag.String("qdrant", "localhost:6334", "qdrant gRPC address")
	ollamaURL  = flag.String("ollama", "http://localhost:11434/api/embed", "ollama embedding URL")
)

func main() {
	ctx := context.Background()
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: obsidian-rag [flags] <command> <vault_path> [question]")
		return
	}

	command := args[0]
	vaultPath := args[1]

	store, err := qdrant.NewQdrantStore(*qdrantAddr)
	if err != nil {
		log.Fatal(err)
	}
	repo := filerepo.NewRepository(os.DirFS(vaultPath))
	parser := markdown.NewMDParser(chunkSize, mergeChunkLimit, minChunkSize)
	embedder := ollama.NewOllamaEmbedder(embedModelName, *ollamaURL)

	engine := domain.NewRagEngine(repo, store, parser, embedder)

	switch command {
	case "index":
		err := engine.Sync(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Index error: %v\n", err)
			os.Exit(1)
		}
	case "ask":
		if len(args) < 3 {
			fmt.Println("Usage: obsidian-rag ask <vault_path> <question>")
			return
		}
		err := engine.Sync(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Index error: %v\n", err)
			os.Exit(1)
		}
		answer, err := engine.Ask(ctx, args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(answer)
	default:
		fmt.Printf("Unknown command: %q\n", command)
	}
}
