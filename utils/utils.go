package utils

import (
	"os/exec"
	"strings"
)

func GetModuleName(dir string) (string, error) {
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	moduleName := strings.TrimSpace(string(output))
	return moduleName, nil
}
