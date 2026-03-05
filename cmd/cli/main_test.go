package main_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/bryack/obsidian_rag/adapters/cli"
	"github.com/bryack/obsidian_rag/specifications"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func TestRAGCLI(t *testing.T) {
	binaryPath, err := ensureBinary()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	container, err := qdrant.Run(ctx, "qdrant/qdrant:latest")
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
		err := os.RemoveAll(filepath.Dir(binaryPath))
		if err != nil {
			t.Logf("Warning: failed to remove temp directory: %v", err)
		}
	})

	grpcEndpoint, err := container.GRPCEndpoint(ctx)
	require.NoError(t, err)

	fakeOllama := setupFakeOllama(t)

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.md")
	os.WriteFile(filePath, []byte("Обсидиан — это база знаний, работающее с локальными Markdown-файлами. Оно позволяет структурировать информацию через гиперссылки."), 0644)

	driver := cli.Driver{
		PathToBinary: binaryPath,
		WorkingDir:   tempDir,
		VaultPath:    tempDir,
		QdrantAddr:   grpcEndpoint,
		OllamaURL:    fakeOllama,
	}
	specifications.RAGSpecification(t, &driver)
}

var (
	buildOnce  sync.Once
	binaryPath string
	buildError error
)

func ensureBinary() (string, error) {
	buildOnce.Do(func() {
		binPath, err := buildBinaryPath()
		if err != nil {
			buildError = err
		}
		binaryPath = binPath
	})
	return binaryPath, buildError
}

func buildBinaryPath() (string, error) {
	tempDir, err := os.MkdirTemp("", "test-binary-*")
	if err != nil {
		return "", fmt.Errorf("failed to make temp directory: %v", err)
	}

	binName := "temp-binary"
	binPath := filepath.Join(tempDir, binName)

	build := exec.Command("go", "build", "-cover", "-o", binPath, ".")
	build.Env = append(os.Environ(), "GOCOVERDIR="+os.Getenv("GOCOVERDIR"))

	var stderr bytes.Buffer
	build.Stderr = &stderr

	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Go build Error: \n%q\n", stderr.String())
		return "", fmt.Errorf("cannot build tool %s: %v", binName, err)
	}

	return binPath, nil
}

func setupFakeOllama(t *testing.T) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vectorStr := strings.Repeat("0,", 1023) + "0"
		fmt.Fprintf(w, "{\"embeddings\": [[%s]]}", vectorStr)
	}))
	t.Cleanup(server.Close)
	return server.URL + "/api/embed"
}
