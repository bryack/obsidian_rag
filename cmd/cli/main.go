package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "ask" {
		fmt.Println("Obsidian RAG: Ответ найден в ваших заметках.")
		return
	}
}
