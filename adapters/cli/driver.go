package cli

import (
	"os/exec"
	"strings"
)

type Driver struct {
	PathToBinary string
	WorkingDir   string
	VaultPath    string
}

func (d *Driver) Ask(question string) (string, error) {
	cmd := exec.Command(d.PathToBinary, "ask", d.VaultPath, question)
	cmd.Dir = d.WorkingDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
