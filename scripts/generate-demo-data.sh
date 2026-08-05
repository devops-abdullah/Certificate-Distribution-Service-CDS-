#!/usr/bin/env bash
#
# Generates a local, throwaway ACME store (real self-signed certs + keys,
# never committed) so `docker compose up` demonstrates the full pipeline:
# manager loads real certs -> exports them -> the agent downloads, installs,
# and reloads Nginx with one of them.
#
# Safe to re-run; skips generation if .demo/acme.json already exists.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEMO_DIR="$ROOT/.demo"
AGENT_DOMAIN="agent-demo.example.com"

mkdir -p "$DEMO_DIR"

if [ -f "$DEMO_DIR/acme.json" ]; then
  echo ".demo/acme.json already exists — skipping generation (delete .demo/ to regenerate)."
  exit 0
fi

gen_cert() {
  local domain="$1" days="$2" out_prefix="$2_$1"
  openssl req -x509 -newkey rsa:2048 \
    -keyout "$DEMO_DIR/$out_prefix.key.pem" -out "$DEMO_DIR/$out_prefix.cert.pem" \
    -noenc -days "$days" -subj "//CN=$domain" -addext "subjectAltName=DNS:$domain" \
    2>/dev/null
  base64_cert=$(openssl base64 -A -in "$DEMO_DIR/$out_prefix.cert.pem")
  base64_key=$(openssl base64 -A -in "$DEMO_DIR/$out_prefix.key.pem")
}

echo "Generating demo certificates..."

gen_cert "new.example.com" 90
new_cert="$base64_cert"; new_key="$base64_key"

gen_cert "expiring.example.com" 4
expiring_cert="$base64_cert"; expiring_key="$base64_key"

gen_cert "$AGENT_DOMAIN" 90
agent_cert="$base64_cert"; agent_key="$base64_key"

cat > "$DEMO_DIR/acme.json" <<JSON
{
  "letsencrypt": {
    "Account": {
      "Email": "admin@example.com",
      "Registration": {"body": {"status": "valid", "contact": []}, "uri": "https://acme.example/acct/1"},
      "PrivateKey": "unused",
      "KeyType": "4096"
    },
    "Certificates": [
      {"domain": {"main": "new.example.com", "sans": []}, "certificate": "$new_cert", "key": "$new_key", "Store": "default"},
      {"domain": {"main": "expiring.example.com", "sans": []}, "certificate": "$expiring_cert", "key": "$expiring_key", "Store": "default"},
      {"domain": {"main": "$AGENT_DOMAIN", "sans": []}, "certificate": "$agent_cert", "key": "$agent_key", "Store": "default"}
    ]
  }
}
JSON

echo "Wrote $DEMO_DIR/acme.json (new.example.com, expiring.example.com, $AGENT_DOMAIN)"
