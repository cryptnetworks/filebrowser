# Security Policy

## Supported Versions

This fork is undergoing security remediation. No patched release is available
yet; fixes are developed on protected branches and will be documented when a
release is published.

The current audit baseline and closure criteria are tracked in the
[security remediation plan](docs/security-remediation-plan.md).

| Version | Supported |
| ------- | --------- |
| Unreleased fork `master` | Security development |
| Existing 2.x releases | ❌ |
| < 2.0   | ❌        |

## Before Reporting

To avoid duplicates, first check the [existing upstream advisories](https://github.com/filebrowser/filebrowser/security/advisories), this fork's security advisories, and open issues.

- **It concerns this project, not a fork.** Reports about code, features, or endpoints that don't exist here belong to the relevant fork.
- **It isn't an already-known class** that remains unaddressed. Those are listed under [Security](README.md#security) in the README; reports covering them are likely to be closed as duplicates.

## Reporting a Vulnerability

Report vulnerabilities privately through this fork's [Security](https://github.com/cryptnetworks/filebrowser/security) page. Do not open a public issue for an undisclosed vulnerability.

Please include, where possible:

- The commit the issue was found at
- A plaintext proof of concept (no binaries)
- Steps to reproduce
- Recommended remediation, if any

Reports are triaged against this fork. A report being accepted does not imply a patched release exists; consult the advisory for affected and patched versions.
