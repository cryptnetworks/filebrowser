# Security remediation plan

Last updated: 2026-08-04

This document tracks the security baseline for `cryptnetworks/filebrowser` and
the work required before a patched fork release can be considered. The current
implementation is under review in [pull request #18](https://github.com/cryptnetworks/filebrowser/pull/18).

## CodeQL baseline

The default-branch scan currently contains 40 open alerts:

| Priority | Alerts | Rule | Remediation status |
| --- | ---: | --- | --- |
| Critical | 1 | `go/command-injection` | PR analysis has zero CodeQL findings; default branch awaits merge and rescan |
| High | 30 | `go/path-injection` | PR analysis has zero CodeQL findings; confinement evidence expanded |
| High | 2 | `go/clear-text-logging` | PR analysis has zero CodeQL findings; sensitive command arguments are not logged |
| High | 1 | `go/incorrect-integer-conversion` | PR analysis has zero CodeQL findings; boundary cases covered |
| Medium | 6 | `actions/missing-workflow-permissions` | PR analysis has zero CodeQL findings; workflow permissions verified |

CodeQL records alerts against the default branch. An alert is not considered
closed merely because a pull request contains a proposed fix.

## Remediation phases

### 1. Land the immediate CodeQL fixes

- Require all command features to use absolute administrator-approved
  executables and explicit argument vectors; never evaluate requests or hook
  configuration through a shell.
- Keep attacker-controlled hook values in single arguments or documented
  environment entries and reject `PATH`-resolved executables.
- Redact authentication secrets from configuration output and remove invalid
  usernames from logs.
- Reject overflowing user IDs instead of accepting a truncated conversion.
- Give workflows explicit read-only permissions and pin third-party actions to
  immutable commit SHAs.
- Restore valid Dependabot coverage for Go, npm, GitHub Actions, and Docker.

Exit criteria: PR #18 passes repository tests, build, lint, and all configured
CodeQL analyses; the protected branch rules remain enabled.

### 2. Verify path confinement findings

The 30 path alerts cross API handlers and filesystem helpers that are intended
to be confined by `ScopedFs` and `BasePathFs`. PR #18 adds API-level tests for
reads, writes, creates, updates, deletes, subtitles, and escaping symlinks, plus
filesystem and settings traversal cases.

After merge, review every surviving alert against the tested confinement
boundary. Fix any demonstrated escape. Dismiss an alert only when all of the
following are recorded in its CodeQL disposition:

- the request is confined before the reported filesystem operation;
- an automated test covers the reported operation and an out-of-scope path;
- symlink behavior is covered where applicable; and
- the alert is a modeling limitation rather than a reachable vulnerability.

The `followExternalSymlinks` option deliberately relaxes confinement and must
remain disabled by default, explicitly documented as unsafe, and excluded from
claims of tenant isolation.

Exit criteria: every path alert is either fixed or individually dispositioned
with test evidence; no bulk dismissal is permitted.

### 3. Address non-CodeQL security debt

- Triage the known upstream advisory classes tracked in issue #2.
- Keep secret scanning and Dependabot alerts at zero.
- Add dependency review and broader supply-chain checks where they do not
  duplicate existing controls.
- Decide which deployment configurations this fork will support and publish a
  hardened configuration baseline.

Exit criteria: each known advisory class has an owner, severity, remediation or
documented mitigation, regression coverage, and a target release.

### 4. Release a patched build

- Re-run race tests, lint, type checking, production builds, CodeQL, dependency
  review, and secret scanning on the release candidate.
- Review security-relevant configuration defaults and upgrade notes.
- Publish a changelog that describes behavior changes without including
  exploitation instructions.
- Update `SECURITY.md` only after a patched artifact is available.

Exit criteria: all required checks pass on the protected default branch, no
critical or high alert remains without an accepted disposition, and the release
artifact is reproducible from the tagged commit.

## Operating rules

- Do not dismiss an alert to make a check pass.
- Do not weaken branch protection or SHA-pinning as a permanent workaround.
- Add a regression test for every confirmed vulnerability.
- Record residual risk and compatibility impact in the pull request that accepts
  it.
- Re-scan the default branch after each security merge and update this document
  when counts or priorities change.

## Pre-merge residual risk

- Command execution, event hooks, and hook authentication remain optional
  privileged features. They are disabled by default and require trusted
  executables plus an external operating-system sandbox when enabled.
- The scoped filesystem performs a resolved-target check immediately before
  each operation. A process that can concurrently replace path components may
  still present a time-of-check/time-of-use race; deployments must prevent
  untrusted local processes from mutating the served tree outside the API.
- `github.com/disintegration/imaging` has an unfixed low-severity crafted-TIFF
  crash advisory. Image processing can be disabled as a deployment mitigation.
- `golang.org/x/crypto/openpgp` is reported as unmaintained, but this application
  reaches `x/crypto` through `bcrypt`, not `openpgp`. The module-level scanner
  result remains documented until the dependency no longer contains that
  package or the scanner supports package reachability.
