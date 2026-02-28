package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/bryack/obsidian_rag/adapters/cli"
	"github.com/bryack/obsidian_rag/specifications"
)

func TestRAGCLI(t *testing.T) {
	binaryPath, err := ensureBinary()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(filepath.Dir(binaryPath))
		if err != nil {
			t.Logf("Warning: failed to remove temp directory: %v", err)
		}
	})

	driver := cli.Driver{
		PathToBinary: binaryPath,
		TempDir:      t.TempDir(),
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
