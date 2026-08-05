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

## Phase 4

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

TLS_CERT

TLS_KEY

TLS_CLIENT_CA
```

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
Milestone 4 / Phase 3 (completed)
```

Current Task:

```
Begin Phase 4: Certificate Agent, Download Engine, Atomic Installation, Nginx Integration
```

Next Tasks:

1. Design the Certificate Agent (runs on customer servers, pulls from the manager)
2. Download engine (agent-side, authenticated against this API)
3. Atomic installation of certs on the agent
4. Nginx reload integration on the agent

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
