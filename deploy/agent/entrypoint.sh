#!/bin/sh
set -e

# Nginx needs a certificate to exist at startup, before the agent has had a
# chance to fetch a real one from the manager. Seed a throwaway self-signed
# placeholder so Nginx can start; the agent overwrites it (and reloads
# Nginx) as soon as it fetches the real certificate.
CERT_DIR="/etc/cds-agent/certs/agent-demo.example.com"
if [ ! -f "$CERT_DIR/fullchain.pem" ]; then
  mkdir -p "$CERT_DIR"
  openssl req -x509 -newkey rsa:2048 -keyout "$CERT_DIR/privkey.pem" -out "$CERT_DIR/fullchain.pem" \
    -noenc -days 1 -subj "/CN=placeholder" 2>/dev/null
fi

nginx -g "daemon off;" &

# Wait for Nginx to actually be up (PID file written) before starting the
# agent, so its first poll doesn't race "nginx -s reload" against Nginx
# still starting.
for _ in $(seq 1 50); do
  [ -f /run/nginx.pid ] && break
  sleep 0.1
done

exec /usr/local/bin/cds-agent
