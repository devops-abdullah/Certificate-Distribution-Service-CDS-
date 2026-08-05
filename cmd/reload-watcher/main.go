// Command reload-watcher runs alongside Nginx (sharing its PID namespace,
// not its container) in deployments where the Certificate Agent and Nginx
// are separate containers. It watches the shared certificate volume the
// agent writes into and reloads Nginx locally whenever something changes,
// since the agent has no way to signal a process in another container.
package main

import (
	"strings"

	"github.com/spf13/viper"

	"github.com/devops-abdullah/cds/internal/reloadwatcher"
	"github.com/devops-abdullah/cds/pkg/logger"
	"github.com/devops-abdullah/cds/pkg/utils"
)

func main() {
	viper.SetDefault("WATCH_DIR", "/etc/cds-agent/certs")
	// -o / --oldest targets Nginx's master process specifically (the first
	// nginx process started), so worker processes never receive SIGHUP
	// directly. Requires only pgrep/pkill (procps) in this image, and a
	// shared PID namespace with the Nginx container — no shared /run
	// volume, no Docker socket, and no changes to the Nginx image at all.
	viper.SetDefault("RELOAD_CMD", "pkill -HUP -o nginx")
	viper.AutomaticEnv()

	watchDir := viper.GetString("WATCH_DIR")
	reloadCmd := viper.GetString("RELOAD_CMD")

	logger.Init("info")

	if strings.TrimSpace(reloadCmd) == "" {
		logger.Log.Fatal("RELOAD_CMD must not be empty")
	}

	logger.Log.
		WithField("watchDir", watchDir).
		WithField("reloadCmd", reloadCmd).
		Info("Reload watcher started")

	w := reloadwatcher.New(watchDir)

	err := w.Start(nil, func() {
		logger.Log.Info("Certificate change detected; reloading")
		if err := utils.RunCommand(reloadCmd); err != nil {
			logger.Log.WithError(err).Warn("Reload command failed")
			return
		}
		logger.Log.Info("Reload succeeded")
	}, func(err error) {
		logger.Log.WithError(err).Warn("Watcher error")
	})
	if err != nil {
		logger.Log.WithError(err).Fatal("Reload watcher failed to start")
	}
}
