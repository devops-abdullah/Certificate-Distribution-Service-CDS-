# Contributing to CDS

Thank you for contributing to the Certificate Distribution Service (CDS).

We welcome bug reports, feature requests, documentation improvements, and code contributions.

---

# Development Workflow

1. Fork the repository (if applicable).
2. Create a feature branch.
3. Implement your changes.
4. Run local quality checks.
5. Update documentation when required.
6. Submit a Pull Request.

---

# Branch Naming

Use descriptive names.

Examples:

feature/acme-parser

feature/export-engine

feature/certificate-agent

bugfix/parser-validation

docs/update-readme

---

# Commit Messages

This repository follows Conventional Commits.

Examples

feat(api): add certificate endpoint

fix(parser): handle invalid JSON

docs(readme): update installation guide

test(acme): add parser tests

---

# Code Style

Before submitting a Pull Request, run:

go fmt ./...

go vet ./...

go test ./...

If configured:

golangci-lint run

---

# Pull Request Checklist

Before requesting review:

* Code builds successfully
* Tests pass
* Documentation updated (if needed)
* No secrets committed
* No certificates committed
* No private keys committed
* New code follows project architecture
* Commit history is clean

---

# Reporting Bugs

Include:

* CDS version
* Go version
* Operating system
* Steps to reproduce
* Expected behavior
* Actual behavior
* Relevant logs

---

# Feature Requests

Feature requests should explain:

* The problem
* Proposed solution
* Alternatives considered
* Operational impact

---

# Security

Do not create public issues for sensitive security vulnerabilities.

Instead, follow SECURITY.md (once available) or contact the maintainers through the designated disclosure process.

Never include:

* Private keys
* Certificates
* API tokens
* Customer information
* Production infrastructure details

---

# Documentation

Documentation is considered part of the codebase.

When adding or changing functionality, update the relevant documentation where appropriate.

---

# Testing Philosophy

Contributors are encouraged to add or update tests alongside new functionality.

Where practical:

* Unit tests
* Integration tests
* Error-path tests

---

# Project Philosophy

The primary goals of CDS are:

* Security
* Reliability
* Maintainability
* Simplicity
* Production readiness

Contributions should align with these principles.

Thank you for helping improve CDS.
