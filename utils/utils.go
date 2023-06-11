package utils

import (
	"os/exec"
	"strings"
)

// GetModuleName returns the name of the Go module for the specified directory.
// This function executes the "go list -m" command in the specified directory
// and returns the output as a string.
// If there is an error executing the command, an error is returned.
// Otherwise, the module name is returned as a string.
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
