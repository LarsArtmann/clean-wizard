# Security Policy

## Supported Versions

`go-finding` is a library, not a long-running service, so "support" means: we
provide fixes for security-relevant bugs in the **latest minor release** only.
Older minor versions do not receive backports.

The public API has been frozen since `v1.0.0`; breaking changes require a major
version bump, so upgrading within the `1.x` line is safe.

## Reporting a Vulnerability

**Do NOT open a public GitHub issue for security vulnerabilities.**

Please report suspected vulnerabilities using **GitHub's private vulnerability
reporting**:

1. Go to the **[Security tab](https://github.com/larsartmann/go-finding/security/advisories/new)**
   and click **"Report a vulnerability"**.
2. Include a clear description and, if possible, a minimal reproduction.
3. You will receive an acknowledgement within **72 hours**.

This keeps the report private to the maintainer and the reporting user until a fix
is coordinated. Until then, please keep the report confidential to avoid exposing
users to exploitation.

## Security Practices

- **Private intake only** — vulnerability reports arrive exclusively through
  GitHub's private vulnerability reporting; public issues are closed and
  redirected here.
- **Acknowledgement within 72 hours** — every report receives a triage
  response within that window, followed by status updates as the fix
  progresses.
- **Fix scope** — security-relevant fixes target the latest minor release;
  the frozen `1.x` API keeps upgrading to the patched release non-breaking.
- **Coordinated disclosure** — fixes land first, then a GitHub Security
  Advisory publishes with credit to the reporter (anonymous credit on
  request). Reporters are asked to hold public details until the advisory
  is out.
- **Dependency hygiene** — `govulncheck` and Dependabot track the dependency
  surface; vulnerabilities already fixed upstream are handled by upgrading.

## Scope

This policy covers the `go-finding` library itself, including the five Go modules
in this repository:

- `github.com/larsartmann/go-finding` (core)
- `github.com/larsartmann/go-finding/pipeline`
- `github.com/larsartmann/go-finding/analysis`
- `github.com/larsartmann/go-finding/toolsdk` (BuildFlow provider SDK)
- `github.com/larsartmann/go-finding/cmd/go-finding` (CLI)

It does **not** cover vulnerabilities in the tools that go-finding wraps (govet,
staticcheck, etc.) — report those to the upstream projects.

## Out of Scope

- Theoretical issues without a concrete exploit path
- Bugs in dependencies that are already fixed upstream (upgrade the dependency)
- Social engineering or phishing

## Disclosure

Once a fix is released, we publish a GitHub Security Advisory crediting the
reporter (unless they prefer to remain anonymous).
