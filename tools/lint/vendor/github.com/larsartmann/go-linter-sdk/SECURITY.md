# Security Policy

## Supported Versions

This library follows semantic versioning. Only the latest released version
receives security fixes; please reproduce a report against `@latest` before
filing it.

| Version        | Supported          |
| -------------- | ------------------ |
| Latest release | :white_check_mark: |
| Older releases | :x:                |

## Reporting a Vulnerability

Do **not** open a public issue for a security problem.

Please report privately via [GitHub security advisories](https://github.com/LarsArtmann/go-linter-sdk/security/advisories/new)
("Report a vulnerability"). Include a description, reproduction steps, and
affected versions if known.

## What to Expect

- This is a personal, best-effort maintained project, so there is no formal
  SLA, but an initial response is committed to within 48 hours of a report.
- You will get an acknowledgement when the report is triaged.
- Confirmed vulnerabilities are fixed in a patch release and credited in the
  changelog (unless you prefer to remain anonymous).

## Security Practices

- Every change is reviewed before it reaches `master`; the auto-commit daemon
  merges only work that has already passed the local quality gate.
- Continuous integration runs static analysis (`golangci-lint`, `go vet`,
  BuildFlow) and the full test suite, including the race detector, on every
  push.
- Dependencies are tracked by GitHub's dependency graph and refreshed with
  BuildFlow; known-vulnerable dependencies are handled via patch releases.
- The library performs no network access and handles no credentials at
  runtime. Linter inputs are read from the caller's working directory, so
  protecting sensitive source paths remains the consumer's responsibility.

## Scope

- The Go library in this repository (`github.com/larsartmann/go-linter-sdk`)
  and the linters built on it.
- Dependencies are tracked by GitHub's dependency graph; known-vulnerable
  dependencies are handled via patch releases.
