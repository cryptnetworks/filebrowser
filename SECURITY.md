# Security policy

## Current support status

This fork is restoring active maintenance. Until the first fork release is published, only the current `master` branch receives security work. No inherited upstream binary or container image should be considered supported by this fork.

| Version | Status |
| --- | --- |
| Current fork `master` | Security hardening in progress |
| Future latest fork release | Planned support |
| Upstream 2.x releases and images | Unsupported by this fork |
| Older releases | Unsupported |

The repository inherits eleven published advisories with no patched upstream version. See [the master tracker](https://github.com/cryptnetworks/filebrowser/issues/2) and [deployment controls](docs/security-maintenance.md).

The current CodeQL audit baseline and alert-closure criteria are tracked in the
[security remediation plan](docs/security-remediation-plan.md).

## Reporting a vulnerability

Use GitHub's **Security → Report a vulnerability** flow for this fork when it is available. Do not file a public issue containing an undisclosed exploit, credentials, private data, or a working proof of concept.

If private vulnerability reporting is temporarily unavailable, open a minimal public issue asking the maintainer to establish a private channel. Include no exploit details beyond the affected component and suspected severity.

A useful private report includes:

- The exact fork commit or release.
- Affected configuration and platform.
- Preconditions and privileges required.
- Minimal plaintext reproduction steps.
- Expected and actual authorization boundaries.
- Impact and suggested remediation.
- Whether the issue is already public or under embargo.

## Response targets

These are operational targets, not guarantees:

- Acknowledge a credible report within two business days.
- Complete initial severity and duplicate assessment within seven days.
- Establish a containment or fix plan for critical findings within fourteen days.
- Coordinate disclosure after a patch or documented mitigation is available.

## Disclosure and fix requirements

A vulnerability is not considered fixed until:

- The vulnerable baseline is reproduced.
- A regression test demonstrates the prior failure.
- The fix passes unit, race, integration, SAST, dependency, and relevant DAST checks.
- Affected and patched versions are documented.
- Upgrade, rollback, and mitigation guidance is available.
- Any temporary suppression has an owner and expiry date.

Security advisories should credit reporters who want attribution and avoid publishing secrets or unnecessary exploit detail.

## Deployment warning

Until the critical and high-priority issues in the master tracker are resolved:

- Disable signup.
- Disable command execution, runners, and hooks.
- Do not accept proxy-auth headers from untrusted networks.
- Use a non-root container with minimal mounts.
- Require a trusted TLS and authentication proxy.
- Maintain tested offline backups.
