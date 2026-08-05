# Certificate Distribution Service (CDS)

> **Production-grade certificate distribution platform for Traefik ACME, Let's Encrypt, ZeroSSL, internal CAs, and future certificate providers.**

---

# Project Vision

The goal of CDS is to eliminate manual SSL certificate management across large-scale deployments.

Instead of every customer server requesting certificates directly from Let's Encrypt, a central Certificate Manager will obtain, monitor, validate and distribute certificates securely to hundreds or thousands of servers.

This reduces:

* Let's Encrypt API requests
* Certificate renewal overhead
* Manual deployments
* Operational complexity
* Risk of rate limiting

---

# Problem Statement

Current workflow:

```
Server 1  ---> Let's Encrypt
Server 2  ---> Let's Encrypt
Server 3  ---> Let's Encrypt
...
Server 300 ---> Let's Encrypt
```

Problems:

* Hundreds of ACME requests
* Rate limit risk
* Difficult certificate management
* No centralized monitoring
* Manual renewals
* Difficult auditing

---

# Proposed Architecture

```
                    Let's Encrypt
                          │
                          ▼
                    Traefik (ACME)
                     acme.json
                          │
                          ▼
         Certificate Distribution Service (Manager)
                          │
              Secure REST API + mTLS
                          │
      ┌───────────────────┼────────────────────┐
      │                   │                    │
      ▼                   ▼                    ▼
 Certificate Agent   Certificate Agent   Certificate Agent
     Customer 1          Customer 2          Customer N
      │                   │                    │
      ▼                   ▼                    ▼
    Nginx               Nginx               Nginx
```

---

# High Level Goals

* Read Traefik ACME certificates
* Support multiple ACME resolvers
* Parse unlimited certificates
* Export certificates securely
* Distribute certificates to remote agents
* Automatically reload Nginx
* Production-grade monitoring
* Production-grade security
* Kubernetes ready
* Docker ready
* Open source

---

# Current Architecture

```
cmd/
    manager/

internal/
    acme/
    api/
    certs/
    config/
    metrics/
    storage/
    logger/
    utils/

examples/
docs/
scripts/
tests/
```

---

# Current Progress

## Milestone 1 - Bootstrap

Status:

✅ Completed

Completed items

* Go module initialized
* Versioned API
* Health endpoint
* Version endpoint
* Configuration package
* Structured logging
* Gin router
* Prometheus metrics endpoint
* Basic project structure
* Successful build
* Docker-ready project layout

---

## Milestone History

### Milestone 2 — ACME Reader

Status: ✅ Completed

* Read ACME file
* Validate JSON
* Parse into Go structures
* Store in memory
* Expose certificate metadata through API (`GET /api/v1/certificates`, `GET /api/v1/certificates/:domain`)

Only certificate metadata (domain, SANs, issuer, serial number, validity window) is exposed. Private keys are never returned by the API.

### Milestone 3 — Phase 2 (Certificate Lifecycle)

Status: ✅ Completed

* **File Watcher** — the ACME file is now watched continuously (`internal/acme/watcher.go`, fsnotify) and the in-memory inventory + on-disk export are reloaded automatically whenever it changes, instead of only at startup.
* **Certificate Validation** — metadata now includes `expiringSoon` (within `CERT_EXPIRY_WARN_DAYS`, default 30) and `notYetValid` alongside `expired`.
* **Certificate Export** — `fullchain.pem`/`privkey.pem` are written per domain under `EXPORT_DIR/<domain>/`, atomically (write-temp-then-rename) with `0644`/`0600` permissions respectively.
* **Storage Layer** — `internal/storage` now exposes a `Repository` interface; the in-memory `Store` is one implementation, so a persistent backend can be added later without touching callers.

### API Dashboard (dev tool)

Status: ✅ Completed

`GET /dashboard` serves a same-origin HTML/JS page that exercises every route the service exposes and reports pass/fail, for checking the API by eye during development instead of hand-typing curl commands. Not part of the public API contract.

### Milestone 4 — Phase 3 (API Security)

Status: ✅ Completed

* **API Authentication** — `API_KEY_READONLY` / `API_KEY_ADMIN` gate the certificate endpoints and the reload endpoint via an `X-API-Key` header (or `Authorization: Bearer`). With neither key configured, the API fails closed (503) rather than serving unauthenticated. `GET /api/v1/health`, `GET /api/v1/version`, `GET /metrics`, and `GET /dashboard` stay public.
* **RBAC** — two roles, `readonly` and `admin` (admin satisfies either requirement). Reading certificate data needs either key; the new `POST /api/v1/reload` endpoint (manually triggers the same reload the file watcher does) needs the admin key.
* **Audit Logging** — every authenticated request logs a structured `"audit": true` entry (method, path, status, role, remote IP, duration) alongside the existing application logs.
* **mTLS** — `TLS_CERT`/`TLS_KEY` enable HTTPS; additionally setting `TLS_CLIENT_CA` requires clients to present a certificate signed by that CA (mutual TLS). With none of the three set, the server runs plain HTTP (local dev, or behind an external TLS terminator).

### Milestone 5 — Phase 4 (Certificate Agent)

Status: ✅ Completed

A second binary, `cmd/agent`, that runs on customer servers:

* **Certificate Agent (core)** — polls the manager on `POLL_INTERVAL_SECONDS` for each domain in `DOMAINS`. Supports mTLS to the manager via `TLS_CLIENT_CERT`/`TLS_CLIENT_KEY`/`TLS_CA`.
* **Download Engine** — fetches `GET /api/v1/certificates/:domain/bundle` (new manager endpoint, gated by a new least-privilege `agent` role/`API_KEY_AGENT` — never the readonly or admin key) and reads the cert + private key it returns.
* **Atomic Installation** — writes `fullchain.pem`/`privkey.pem` to `INSTALL_DIR/<domain>/` the same write-temp-then-rename way the manager's exporter does, and detects whether the content actually changed so unchanged polls don't trigger a reload.
* **Nginx Integration** — runs `NGINX_RELOAD_CMD` (default `nginx -s reload`) only when a certificate actually changed.

The bundle endpoint reads whatever the manager's exporter already wrote to `EXPORT_DIR`, so Milestone 3's Certificate Export directly backs Milestone 5's Download Engine — no separate storage was added for raw key material.

`Dockerfile.agent` builds the agent alongside Nginx in one image (they run on the same host in a real deployment, since the agent needs to signal Nginx); see `deploy/agent/` for the reference Nginx config and entrypoint.

---

# Planned Milestones

## Phase 1 — done

* Bootstrap
* ACME Reader
* ACME Parser
* Certificate Inventory API

## Phase 2 — done

* Certificate Export
* File Watcher
* Storage Layer
* Certificate Validation

## Phase 3 — done

* API Authentication
* mTLS
* RBAC
* Audit Logging

## Phase 4 — done

* Certificate Agent
* Download Engine
* Atomic Installation
* Nginx Integration

## Phase 5

* Prometheus Metrics
* Grafana Dashboards
* Alerting

## Phase 6

* Docker
* Kubernetes
* GitHub Actions
* Documentation
* Production Release

---

# Repository Standards

## Branch Strategy

Never commit directly to `main`.

Example:

```
main

develop

feature/acme-reader

feature/acme-parser

feature/export-engine

feature/certificate-agent
```

---

## Commit Convention

Use Conventional Commits.

Examples

```
feat(api): add health endpoint

feat(acme): implement parser

fix(export): handle empty certificate

docs(readme): update roadmap

refactor(storage): simplify cache
```

---

## Pull Request Rules

Every Pull Request must:

* Build successfully
* Pass all tests
* Pass `go vet`
* Pass `golangci-lint`
* Pass security scans
* Build Docker image
* Update documentation when applicable

---

# GitHub Repository Standards

Repository uses:

* GitHub Projects
* GitHub Milestones
* GitHub Labels
* GitHub Issues
* Pull Requests
* GitHub Actions

Development workflow:

Issue

↓

Feature Branch

↓

Commit

↓

Pull Request

↓

CI Validation

↓

Code Review

↓

Merge

---

# Quality Gates

No Pull Request may be merged unless:

* CI passes
* Go build passes
* Go tests pass
* gofmt passes
* go vet passes
* golangci-lint passes
* govulncheck passes
* Trivy passes
* Gitleaks passes
* Docker build succeeds
* Documentation updated (when applicable)

---

# Security Principles

* Never commit `acme.json`
* Never commit certificates
* Never commit private keys
* Never hardcode domains
* Never hardcode secrets

Repository uses example domains only:

```
example.com

*.example.com
```

---

# Configuration

Everything must be configurable through environment variables.

Examples

```
PORT

LOG_LEVEL

ACME_FILE

EXPORT_DIR

CERT_EXPIRY_WARN_DAYS

API_KEY_READONLY

API_KEY_ADMIN

API_KEY_AGENT

TLS_CERT

TLS_KEY

TLS_CLIENT_CA
```

## Certificate Agent Configuration (`cmd/agent`)

The agent is a separate binary/process with its own environment variables:

```
MANAGER_URL

API_KEY

DOMAINS

POLL_INTERVAL_SECONDS

INSTALL_DIR

NGINX_RELOAD_CMD

TLS_CLIENT_CERT

TLS_CLIENT_KEY

TLS_CA
```

## Try the Full Deployment (Manager + Agent + Nginx)

```bash
./scripts/generate-demo-data.sh   # generates .env (random API keys) and ./.demo (real, throwaway certs) — neither is ever committed
docker compose up -d --build
docker compose logs -f cds agent  # watch both services live
```

Both the certs and the API keys are generated locally by that script and gitignored — nothing key-shaped ever lives in this repo's tracked files or history, so a secret scanner has nothing to (correctly or incorrectly) flag.

This starts three containers:

* `export-init` — a one-shot container that fixes permissions on the `export-data` volume so the manager (which runs as a non-root user) can write to it, then exits. Without this, exporting a *real* certificate fails with a permission error the first time the volume is created (the redacted placeholder keys in `examples/acme.sample.json` never hit this, since they fail to decode before ever reaching the filesystem write).
* `cds` — the manager, on `:8080`, loaded with 3 demo certificates (one expiring in ~4 days, to show `expiringSoon` in action).
* `agent` — the Certificate Agent + Nginx, on `:8443`. It polls the manager every 10s for `agent-demo.example.com`, installs the certificate it gets back, and reloads Nginx only when the certificate actually changed.

To see it prove itself end to end:

```bash
# Nginx is serving the real certificate the manager handed the agent, not a placeholder:
# (uses openssl rather than curl -v — Windows' bundled curl uses the Schannel
# TLS backend, whose -v output has no "subject:" line at all, unlike
# OpenSSL-backed curl builds on Linux/macOS; openssl s_client behaves the
# same everywhere)
echo | openssl s_client -connect localhost:8443 2>/dev/null | openssl x509 -noout -subject

# Full certificate inventory, including the expiring one (reads the key .env just generated):
curl -s -H "X-API-Key: $(grep API_KEY_READONLY .env | cut -d= -f2)" http://localhost:8080/api/v1/certificates
```

Tear down with `docker compose down` (add `-v` to also drop the `export-data` volume).

`API_KEY` should be the manager's `API_KEY_AGENT` value — never the readonly or admin key. `DOMAINS` is a comma-separated list; the agent only ever learns about the domains it's explicitly configured for.

## Nginx Already Deployed Separately? (Separate Containers)

`Dockerfile.agent` bundles the agent with Nginx because `nginx -s reload` signals a *local* process — that only works when they share a container. If Nginx already runs on its own (its own container, its own image you don't want to touch), use this topology instead:

* **`Dockerfile.agent-only`** — the agent alone, no Nginx bundled. Just fetches from the manager and writes into a shared volume.
* **`Dockerfile.reload-watcher`** (`cmd/reload-watcher`) — a tiny separate process that watches that same shared volume and reloads Nginx when something changes. It runs in **its own container** but shares Nginx's **PID namespace** (`pid: "service:nginx"` in Compose, or the same Pod in Kubernetes), so it can send `nginx` a `SIGHUP` directly — no Docker socket access, no shared `/run` volume, and **no changes to your existing Nginx image at all**.

```
agent (own container) --writes--> [shared cert volume] <--reads-- nginx (your existing, untouched image)
                                                              ^
                                          reload-watcher (own container, shares nginx's PID namespace)
```

Try it:

```bash
./scripts/generate-demo-data.sh
docker compose -f docker-compose.separate.yml -p cds-separate up -d --build
docker compose -f docker-compose.separate.yml -p cds-separate logs -f agent reload-watcher
```

This starts `cds` (manager), `agent` (standalone), a stock `nginx:alpine` (standing in for your own already-deployed Nginx, completely unmodified), and `reload-watcher` alongside it.

Note: `NGINX_RELOAD_CMD` has no default — it's opt-in, not assumed. Set it on the agent only if it's the one with Nginx access (the combined `Dockerfile.agent` topology); leave it unset when a separate `reload-watcher` handles reload instead.

Tear down with `docker compose -f docker-compose.separate.yml -p cds-separate down -v`.

---

# Future Providers

CDS should not be tightly coupled to Traefik.

Future providers should include:

* Traefik
* Let's Encrypt
* ZeroSSL
* HashiCorp Vault
* cert-manager
* Internal CA
* AWS ACM
* Azure Key Vault

---

# Long-Term Goals

* High Availability
* Multi-tenancy
* Web UI
* gRPC API
* WebSocket Notifications
* Cluster Support
* Automatic Agent Registration
* Plugin Architecture

---

# Engineering Principles

This project follows several principles:

1. Every commit must build successfully.
2. Every feature is developed in a dedicated branch.
3. Every Pull Request must pass all quality gates.
4. Every new feature should include documentation updates when needed.
5. Security takes priority over convenience.
6. Avoid hardcoded values and company-specific information.
7. Keep components modular so additional certificate providers can be added without major refactoring.

---

# Development Workflow

For every implementation:

1. Create GitHub Issue
2. Create feature branch
3. Implement feature
4. Run tests
5. Update documentation
6. Open Pull Request
7. CI validation
8. Review
9. Merge

---

# Current Status

Current Version:

```
v0.1.0
```

Current Milestone:

```
Milestone 5 / Phase 4 (completed)
```

Current Task:

```
Begin Phase 5: Prometheus Metrics, Grafana Dashboards, Alerting
```

Next Tasks:

1. Agent-side metrics (poll success/failure, install/reload counts)
2. Grafana dashboards for manager + agent fleet visibility
3. Alerting on expiring/failed certificates and agent staleness

---

# Notes for AI Development

This repository is intended to be developed with AI-assisted engineering tools (Claude Code, ChatGPT, Codex, etc.).

Requirements:

* Always maintain production-quality code.
* Follow Go best practices.
* Keep packages small and focused.
* Avoid introducing breaking changes without justification.
* Prefer clear abstractions over tightly coupled implementations.
* Keep the project provider-agnostic so additional certificate backends can be added in the future.
* Always ensure `go build`, `go test`, and `go vet` succeed before considering a milestone complete.
