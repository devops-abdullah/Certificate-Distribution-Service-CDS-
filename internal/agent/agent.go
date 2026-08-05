package agent

import "time"

// LogFunc receives a structured event name plus fields, letting Run stay
// decoupled from any particular logging library.
type LogFunc func(event string, fields map[string]interface{})

// Run polls the manager for each configured domain on cfg.PollInterval,
// installing and reloading whenever a certificate's content changes. It
// runs one poll immediately, then blocks until stop is closed.
func Run(cfg Config, client *Client, installer *Installer, onLog LogFunc, stop <-chan struct{}) {
	poll(cfg, client, installer, onLog)

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			poll(cfg, client, installer, onLog)
		}
	}
}

func poll(cfg Config, client *Client, installer *Installer, onLog LogFunc) {
	for _, domain := range cfg.Domains {
		bundle, err := client.FetchBundle(domain)
		if err != nil {
			onLog("fetch_failed", map[string]interface{}{"domain": domain, "error": err.Error()})
			continue
		}

		changed, err := installer.Install(bundle)
		if err != nil {
			onLog("install_failed", map[string]interface{}{"domain": domain, "error": err.Error()})
			continue
		}

		if !changed {
			onLog("unchanged", map[string]interface{}{"domain": domain})
			continue
		}

		onLog("installed", map[string]interface{}{"domain": domain})

		if err := ReloadNginx(cfg.NginxReloadCmd); err != nil {
			onLog("reload_failed", map[string]interface{}{"domain": domain, "error": err.Error()})
			continue
		}

		onLog("reloaded", map[string]interface{}{"domain": domain})
	}
}
