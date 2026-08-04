# Security maintenance for this fork

This fork is being brought under active maintenance after the upstream project announced its wind-down. Current builds must still be treated as security-hardening work in progress. Do not expose them directly to the public internet without a trusted TLS-terminating authentication proxy.

The master security tracker is [issue #2](https://github.com/cryptnetworks/filebrowser/issues/2). Implementation issues #3 through #15 and #17 split the critical vulnerabilities, scanning, documentation, inherited dependency findings, and functional backlog into reviewable work.

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

## Filesystem isolation

Every user filesystem is rooted below the configured server root. With the
default `followExternalSymlinks=false`, each filesystem operation first checks
the resolved target and rejects symbolic links that lead outside the authorized
root. This covers reads, writes, directory creation, TUS uploads, copy, move,
rename, archive generation, and deletion through the shared filesystem layer.

Setting `followExternalSymlinks=true` replaces that boundary with lexical path
confinement and deliberately permits links to targets outside the user scope.
It is not a tenant-isolation mode. Enable it only when the operator owns and
trusts every reachable link target, records the risk acceptance, and constrains
the process with a minimal operating-system mount. It defaults to disabled for
new and existing configurations unless explicitly enabled.

## Security automation

| Layer | Control | When it runs | Initial policy |
| --- | --- | --- | --- |
| SAST | GitHub-managed CodeQL default setup | Pull requests, protected-branch pushes, weekly | Extended queries for GitHub Actions, Go, and JavaScript/TypeScript; high-or-higher security alerts fail |
| Dependency diff | `Security - Dependency Review` | Pull requests | Reject newly introduced high/critical vulnerable dependencies |
| Dependency inventory | `Security - OSV Scanner` | Pull requests, merge queue, master, weekly | Block newly introduced PR findings and publish the complete scheduled baseline |
| DAST | `Security - DAST Baseline` | Relevant pull requests, weekly, manual | Passive OWASP ZAP scan of an ephemeral noauth instance with retained JSON, HTML, and Markdown evidence |
| Secrets | GitHub secret scanning and push protection | Every push | Block supported secrets; review and revoke every valid finding |
| Artifact/container (“MAST”) | Tracked in issue #13 | Release pipeline | SBOM, malware, vulnerability, signature, and provenance gates |

This repository has no native mobile application, so “MAST” is treated as malware, artifact, and supply-chain testing. If native mobile code is introduced, add platform-specific mobile application security testing.

DAST active attacks must target only ephemeral CI instances. Authenticated contexts and API scans are tracked in issue #12.

## Scanner and action provenance

Security tools are part of the attack surface. GitHub Actions policy requires full-length commit-SHA pins. Repository workflows also pin the ZAP container by digest, grant minimum token permissions, set timeouts, cancel stale runs, use lockfile caches, and retain security evidence for review. Dependabot owns reviewed updates to action revisions.

The GHCR publishing workflow builds the standard and S6 images for supported platforms, records BuildKit SBOM and provenance attestations, and emits commit-addressable `sha-*` tags. Treat `latest` and `s6` as rolling development channels; production deployments should use a release tag and pin its registry digest.

A March 2026 compromise affected Trivy releases and action tags. Do not add Trivy through a mutable tag. Before adopting it or an alternative, verify the release signature, immutable commit/image digest, incident-safe version, and the permissions/network access it receives.

## Finding triage

1. Confirm the finding against the current commit with a minimal reproduction.
2. Determine reachability and the least privileges required.
3. Link it to a public issue only after disclosure is safe.
4. Critical and high findings block release unless an owner records a time-limited risk acceptance.
5. Suppressions must name an owner, reason, expiry date, and tracking issue.
6. Add a regression test before marking a vulnerability fixed.
7. Publish affected and patched versions when a release is available.

Inherited dependency findings are owned by [issue #17](https://github.com/cryptnetworks/filebrowser/issues/17). Differential PR scans remain blocking for newly introduced vulnerabilities; the full baseline stays visible until remediation makes it suitable as a required release gate.

## Required release evidence

A security-ready release should include:

- Passing unit, race, frontend, integration, SAST, dependency, and DAST checks.
- A machine-readable SBOM for the binary and container.
- Checksums, a verifiable signature, and build provenance.
- A vulnerability summary with explicit accepted risks.
- Upgrade, backup, rollback, and incident-response instructions.
