package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunCommand runs an operator-configured shell command (e.g. an Nginx
// reload command) and returns its combined output on failure. An empty
// command is a no-op — useful in tests, or when the action is handled by
// something else entirely.
func RunCommand(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}

	fields := strings.Fields(command)

	cmd := exec.Command(fields[0], fields[1:]...) //nolint:gosec // command comes from operator-supplied config, not request input

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %q failed: %w (output: %s)", command, err, output)
	}

	return nil
}
