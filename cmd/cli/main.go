package main

import (
	"fmt"
	"os"

	"github.com/bryack/obsidian_rag/internal/domain"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "ask" {
		engine := domain.NewRagEngine()
		answer, err := engine.Ask(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(answer)
		return
	}
}
