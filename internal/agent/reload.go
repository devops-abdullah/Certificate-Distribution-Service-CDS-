package agent

import "github.com/devops-abdullah/cds/pkg/utils"

// ReloadNginx runs the configured reload command (e.g. "nginx -s reload").
// An empty command is a no-op — useful in tests, or when reload is handled
// by something else entirely (e.g. a separate reload-watcher when Nginx
// runs in its own container).
func ReloadNginx(command string) error {
	return utils.RunCommand(command)
}
