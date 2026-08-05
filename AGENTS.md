# AGENTS.md

# AI Development Guide

This document defines the engineering standards, architectural principles, and implementation rules for all AI coding assistants working on the Certificate Distribution Service (CDS).

Supported assistants include (but are not limited to):

* Claude Code
* ChatGPT
* Codex
* GitHub Copilot
* Cursor AI

---

# Project Mission

Build a production-grade Certificate Distribution Service capable of securely distributing TLS certificates from centralized providers to customer infrastructure.

The project must prioritize:

* Reliability
* Security
* Maintainability
* Simplicity
* Extensibility

---

# Current Project Status

Current Version

v0.1.0

Current Milestone

Milestone 5 / Phase 4 — ✅ Completed

Current Objective

Phase 5: Prometheus Metrics, Grafana Dashboards, Alerting.

Milestone 2 (ACME Reader), Milestone 3 (Phase 2: File Watcher, Certificate Validation, Certificate Export, Storage Layer), Milestone 4 (Phase 3: API Authentication, RBAC, Audit Logging, mTLS), and Milestone 5 (Phase 4: Certificate Agent, Download Engine, Atomic Installation, Nginx Integration) are all done:

* `internal/acme` parses the Traefik ACME store into Go structs and watches it (`fsnotify`) for live reload.
* `internal/certs` decodes certificate metadata (domain, SANs, issuer, serial, expiry, `expiringSoon`, `notYetValid` — never key material for the API) and, separately, exports `fullchain.pem`/`privkey.pem` per domain under `EXPORT_DIR` atomically with `0644`/`0600` permissions.
* `internal/storage` exposes a `Repository` interface; the in-memory `Store` is one implementation.
* `internal/api` authenticates via `API_KEY_READONLY`/`API_KEY_ADMIN`/`API_KEY_AGENT` (fails closed if none are set), enforces RBAC (`readonly`/`admin`/`agent`, admin satisfies any check), and audit-logs every authenticated request (including failed attempts).
* `internal/server` builds the TLS config `cmd/server/main.go` serves with; setting `TLS_CLIENT_CA` upgrades it to mTLS.
* `GET /api/v1/certificates` / `GET /api/v1/certificates/:domain` expose metadata only, never private keys. `GET /api/v1/certificates/:domain/bundle` (agent/admin only) serves the actual cert+key, read from `EXPORT_DIR`. `POST /api/v1/reload` (admin-only) triggers an immediate re-read of the ACME store.
* `GET /dashboard` is a dev-only same-origin page that exercises every route and reports pass/fail.
* `cmd/agent` (`internal/agent`) is the Certificate Agent: polls the manager's bundle endpoint per configured domain, installs atomically via `pkg/utils.WriteFileAtomic` (shared with the manager's exporter), skips reload when content is unchanged, and runs `NGINX_RELOAD_CMD` when it does change (opt-in — no default). `Dockerfile.agent` + `deploy/agent/` build it alongside Nginx for a co-located deployment. For a separate-containers deployment (Nginx already exists, untouched), use `Dockerfile.agent-only` (no Nginx bundled) plus `cmd/reload-watcher`/`Dockerfile.reload-watcher` — a tiny process that shares Nginx's PID namespace and signals it directly when the shared cert volume changes, without a Docker socket or any change to the Nginx image. See `docker-compose.separate.yml`.

---

# Development Principles

Always prefer:

* Small packages
* Small functions
* Clear interfaces
* Dependency injection where appropriate
* Standard library before external dependencies
* Context-aware operations
* Structured logging
* Explicit error handling

Avoid:

* Global variables
* Hardcoded paths
* Hardcoded domains
* Panic in production code
* Hidden side effects

---

# Domain Policy

Never use company domains.

Never reference customer infrastructure.

Always use:

example.com

*.example.com

localhost

127.0.0.1

---

# Repository Structure

Keep packages organized.

cmd/

manager/

agent/

internal/

acme/

api/

config/

storage/

metrics/

logger/

certs/

auth/

pkg/

docs/

examples/

scripts/

tests/

---

# Coding Standards

Use Go best practices.

Always:

* run gofmt
* keep files focused
* return errors
* document exported functions
* avoid duplicated logic

---

# Logging

Use structured logging.

Never use:

fmt.Println()

for production logging.

---

# Configuration

Everything must be configurable.

Never hardcode:

ports

paths

domains

API keys

certificate locations

Use environment variables or configuration files.

---

# Security Rules

Never commit:

* acme.json
* certificates
* private keys
* secrets
* passwords
* API tokens

Never expose private keys through the API.

---

# Certificate Sources

Design CDS to support multiple providers.

Current provider

Traefik ACME

Future providers

* Let's Encrypt
* ZeroSSL
* HashiCorp Vault
* cert-manager
* AWS ACM
* Azure Key Vault

The internal API should not depend on provider-specific formats.

---

# Git Workflow

Every feature:

Issue

↓

Feature Branch

↓

Commit

↓

Pull Request

↓

CI

↓

Review

↓

Merge

Never commit directly to main.

---

# Commit Format

Use Conventional Commits.

Examples

feat(acme): implement parser

fix(api): validate request

docs(readme): update roadmap

refactor(storage): simplify cache

---

# Pull Requests

Every Pull Request should:

Build successfully.

Pass tests.

Pass lint.

Pass go vet.

Pass security scans.

Update documentation if required.

---

# Testing

Every new package should include tests where practical.

Run:

go test ./...

before completing work.

---

# Documentation

Update documentation whenever behavior changes.

README.md

Architecture

API

Configuration

Deployment

---

# Long-Term Goals

Support:

* HA
* Kubernetes
* Docker
* Multiple certificate providers
* Agent clustering
* Web UI
* gRPC
* Plugin architecture

Always keep future extensibility in mind.
