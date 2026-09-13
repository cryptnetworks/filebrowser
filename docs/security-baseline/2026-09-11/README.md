# Review evidence, 2026-09-11

Owner: `cryptnetworks`; dependency tracker: #17; CI tracker: #10.

- `osv-results.json` and `osv-results.sarif`: unmodified `osv-full-results`
  artifact 10282762778 from workflow run 34649423810 on master commit
  `1e232ed201a2287cdc9d183274354ad7b05d53d1`.
- `dependabot-open.json` and `codeql-open.json`: read-only GitHub REST exports
  from `cryptnetworks/filebrowser`, `state=open`, obtained during this review.
  Both contain zero alerts; they are not a replacement for OSV results.
- `govulncheck.txt`: govulncheck v1.8.0 run against the local first-batch working
  tree (unchanged Go module manifest) with Go 1.26.6 on macOS arm64.
- `ci-timings.json`: five most recent successful master CI runs from the API,
  before this change. Median elapsed run time is 99 seconds. This uses run
  timestamps, not billed Actions minutes. After-change GitHub timing and billed
  minutes are not yet available; no speedup is claimed.

Reproduce exports using `gh api` for repository `dependabot/alerts?state=open`,
`code-scanning/alerts?state=open`, and the relevant Actions run/artifact endpoints.
Use pagination if an endpoint contains more than one page. Download the artifact
for the exact commit; do not compare unrelated scans or treat expired evidence
as a fresh scan. The exception rationale and review deadline are in the
[security remediation plan](../../security-remediation-plan.md).
