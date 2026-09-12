# Security remediation plan

Last reviewed: 2026-09-11. Owner: `cryptnetworks`. Review monthly, after security
merges, and before release. Track work in [#2](https://github.com/cryptnetworks/filebrowser/issues/2)
and dependency findings in [#17](https://github.com/cryptnetworks/filebrowser/issues/17).

## Current evidence

The reviewed master commit is `1e232ed201a2287cdc9d183274354ad7b05d53d1`.
The [OSV workflow run](https://github.com/cryptnetworks/filebrowser/actions/runs/34649423810)
completed successfully but its full report still contains one low-severity
finding. Open Dependabot and code-scanning API exports returned zero alerts.
These are different inventories; an empty GitHub alert list is not a clean OSV scan.

Raw evidence is retained under [security-baseline/2026-09-11](security-baseline/2026-09-11/README.md):
OSV JSON/SARIF, open alert exports, local `govulncheck` output, and CI timing data.
No compatible dependency upgrade is indicated by the remaining report.

| Advisory aliases | Dependency / exposure | Treatment |
| --- | --- | --- |
| GHSA-q7pp-wcgr-pffx / CVE-2023-36308 | Direct Go dependency `github.com/disintegration/imaging` 1.6.2; shipped preview processing; low severity; no fixed version in report | The crafted TIFF palette condition is rejected before imaging transforms. Keep the scanner finding visible; `img/service_test.go` covers malformed palettes. This is mitigation, not a patched dependency. |
| GO-2026-5932 | Direct module `golang.org/x/crypto`; application imports bcrypt, not affected OpenPGP packages | Existing `osv-scanner.toml` exception retained. Recheck package reachability when imports change. Local govulncheck reports no reachable vulnerability. |

Both dispositions were previously recorded by `cryptnetworks` on 2026-08-04.
Review is due **2026-11-02**, or sooner if code, dependencies, or the trust boundary
changes. Issue #17 owns review. Limit preview resource exposure and retain
malformed-image regression tests for imaging; prohibit introducing OpenPGP
without reassessing the exception. Expiry is a manual review deadline, not an
automatically enforced OSV setting.

## Implemented hardening and remaining work

PR #18 and subsequent security PRs are merged. The old 40-alert CodeQL table was
a historical baseline, not the current status. Implemented work includes scoped
filesystem checks, command argument handling, authentication-secret redaction,
integer boundary checks, TUS length/integrity fixes, and share response redaction.
These changes do not complete every advisory workstream.

- **Proxy authentication (#4):** trusted immediate-peer CIDRs and duplicate-header
  rejection exist. The first-batch change adds explicit default-off provisioning,
  case-variant/merged-header rejection, policy tests, and proxy deployment examples.
- **Sessions (#5):** JWT revocation and single-use refresh remain architectural work.
- **Execution (#6):** optional commands/hooks still require an external OS sandbox.
  Argument parsing does not provide process, memory, network, or filesystem isolation.
- **Paths (#7):** scoped filesystem and recursive authorization tests exist. A local
  process racing path-component replacement can still present a TOCTOU risk.
  Prevent untrusted external writers; keep `followExternalSymlinks=false`.
- **Uploads (#8):** complete configurable budgets and interrupted-upload scenarios.
- **Shares (#9):** complete short-lived authorization sessions and revocation.
- **CI/scanning (#10–#13):** preserve check names and pinning, verify enforcement on
  GitHub, extend authenticated DAST, and finish release artifact signing/scanning.
- **Docs/backlog (#14–#15):** keep claims tied to shipped versions and reproduce
  inherited reports against the current fork before declaring them fixed.

## Scanner enforcement

PR OSV comparison blocks new findings. The first-batch change selects the exact
PR/merge-group base commit and requires structurally valid JSON reports before
comparison, so a failed checkout or missing/malformed report cannot silently
become a baseline. A valid report still requires review for scanner coverage.
Full scans retain JSON/SARIF and remain nonblocking for vulnerability findings;
the existing imaging finding is not newly suppressed by this change.

Promote the full scan only after removing the finding or implementing an explicitly
reviewed, expiring exception gate that preserves the full inventory. GitHub branch
protection currently requires CI and CodeQL checks; dependency/OSV enforcement
must also be verified before claiming a required release gate. #17 remains open
until that evidence and gate are complete.

## Release criteria

Use the [operations runbook](operations.md#release). Require unit/integration/race
tests, frontend lint/tests/type checking, production builds, current CodeQL and
dependency evidence, and relevant DAST. Critical/high findings need remediation
or an explicit, time-limited disposition. Publish affected and patched versions
only when a verified artifact exists. Never dismiss alerts or weaken protection
merely to make a check pass.
