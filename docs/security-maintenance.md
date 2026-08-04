# Security maintenance for this fork

This fork is being brought under active maintenance after the upstream project announced its wind-down. Current builds must still be treated as security-hardening work in progress. Do not expose them directly to the public internet without a trusted TLS-terminating authentication proxy.

The master security tracker is [issue #2](https://github.com/cryptnetworks/filebrowser/issues/2). Implementation issues #3 through #15 split the critical vulnerabilities, scanning, documentation, and inherited backlog into reviewable work.

## Immediate deployment controls

Until the linked fixes ship:

- Keep self-signup disabled.
- Keep command execution, runners, and hooks disabled.
- Do not trust proxy-auth headers from arbitrary clients.
- Run File Browser as a non-root user in a container.
- Mount only the directory that File Browser is intended to serve.
- Put the service behind a reverse proxy that supplies TLS and independent authentication.
- Assume a stolen JWT remains usable until it expires.
- Maintain tested offline backups before enabling uploads, delete, rename, or move.

## Security automation

| Layer | Workflow | When it runs | Initial policy |
| --- | --- | --- | --- |
| SAST | `Security - CodeQL` | Pull requests, merge queue, master, weekly | Go and JavaScript/TypeScript security-and-quality queries |
| Dependency diff | `Security - Dependency Review` | Pull requests | Reject newly introduced high/critical vulnerable dependencies |
| Dependency inventory | `Security - OSV Scanner` | Pull requests, merge queue, master, weekly | Differential PR scan and complete scheduled scan |
| DAST | `Security - DAST Baseline` | Relevant pull requests, weekly, manual | Passive OWASP ZAP scan of an ephemeral noauth instance |
| Secrets | GitHub secret scanning | Every push | Review and revoke every valid finding |
| Artifact/container | Tracked in issue #13 | Release pipeline | SBOM, malware, vulnerability, signature, and provenance gates |

DAST active attacks must target only ephemeral CI instances. Authenticated contexts and API scans are tracked in issue #12.

## Scanner and action provenance

Security tools are part of the attack surface. Prefer GitHub-owned actions or upstream reusable workflows, grant minimum token permissions, set timeouts, and pin third-party actions or images to reviewed immutable revisions.

A March 2026 compromise affected Trivy releases and action tags. Do not add Trivy through a mutable tag. Before adopting it or an alternative, verify the release signature, immutable commit/image digest, incident-safe version, and the permissions/network access it receives.

## Finding triage

1. Confirm the finding against the current commit with a minimal reproduction.
2. Determine reachability and the least privileges required.
3. Link it to a public issue only after disclosure is safe.
4. Critical and high findings block release unless an owner records a time-limited risk acceptance.
5. Suppressions must name an owner, reason, expiry date, and tracking issue.
6. Add a regression test before marking a vulnerability fixed.
7. Publish affected and patched versions when a release is available.

## Required release evidence

A security-ready release should include:

- Passing unit, race, frontend, integration, SAST, dependency, and DAST checks.
- A machine-readable SBOM for the binary and container.
- Checksums, a verifiable signature, and build provenance.
- A vulnerability summary with explicit accepted risks.
- Upgrade, backup, rollback, and incident-response instructions.
