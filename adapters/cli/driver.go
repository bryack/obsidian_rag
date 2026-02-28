package cli

import (
	"os/exec"
	"strings"
)

type Driver struct {
	PathToBinary string
	TempDir      string
}

func (d *Driver) Ask(question string) (string, error) {
	cmd := exec.Command(d.PathToBinary, "ask", question)
	cmd.Dir = d.TempDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
