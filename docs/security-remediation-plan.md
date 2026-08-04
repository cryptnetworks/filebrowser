# Security remediation plan

Last updated: 2026-08-04

This document tracks the security baseline for `cryptnetworks/filebrowser` and
the work required before a patched fork release can be considered. The current
implementation is under review in [pull request #18](https://github.com/cryptnetworks/filebrowser/pull/18).

## CodeQL baseline

The default-branch scan currently contains 40 open alerts:

| Priority | Alerts | Rule | Remediation status |
| --- | ---: | --- | --- |
| Critical | 1 | `go/command-injection` | Fixed in PR #18; awaiting CodeQL verification |
| High | 30 | `go/path-injection` | Confinement tests added; verify each result after merge |
| High | 2 | `go/clear-text-logging` | Fixed in PR #18; awaiting CodeQL verification |
| High | 1 | `go/incorrect-integer-conversion` | Fixed in PR #18; awaiting CodeQL verification |
| Medium | 6 | `actions/missing-workflow-permissions` | Fixed in PR #18; awaiting CodeQL verification |

CodeQL records alerts against the default branch. An alert is not considered
closed merely because a pull request contains a proposed fix.

## Remediation phases

### 1. Land the immediate CodeQL fixes

- Require interactive commands to match an administrator-approved argument
  vector exactly; never evaluate the request through a shell.
- Keep attacker-controlled hook values out of shell command text and pass them
  through the process environment or direct argument vectors.
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
