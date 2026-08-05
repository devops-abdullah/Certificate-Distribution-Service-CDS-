package agent

import (
	"fmt"
	"os/exec"
	"strings"
)

// ReloadNginx runs the configured reload command (e.g. "nginx -s reload").
// An empty command is a no-op — useful in tests, or when reload is handled
// by something else entirely (e.g. an external orchestrator).
func ReloadNginx(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}

	fields := strings.Fields(command)

	cmd := exec.Command(fields[0], fields[1:]...) //nolint:gosec // command comes from operator-supplied config, not request input

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reload command %q failed: %w (output: %s)", command, err, output)
	}

	return nil
}
