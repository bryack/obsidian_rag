package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

type Driver struct {
	PathToBinary string
	WorkingDir   string
	VaultPath    string
	QdrantAddr   string
	OllamaURL    string
}

func (d *Driver) Ask(question string) (string, error) {
	cmd := exec.Command(d.PathToBinary,
		"-qdrant", d.QdrantAddr,
		"-ollama", d.OllamaURL,
		"ask", d.VaultPath, question)
	cmd.Dir = d.WorkingDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

func (d *Driver) Index() error {
	cmd := exec.Command(d.PathToBinary,
		"-qdrant", d.QdrantAddr,
		"-ollama", d.OllamaURL,
		"index", d.VaultPath)
	cmd.Dir = d.WorkingDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}

	return nil
}
