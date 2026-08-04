> [!WARNING]
>
> This fork is under active security maintenance. It inherits known unpatched vulnerabilities from upstream and does not yet have a security-ready release. Follow the [hardening guidance](docs/security-maintenance.md) and do not expose it directly to the public internet.

<p align="center">
  <img src="./branding/banner.png" width="550"/>
</p>

File Browser provides a file-managing interface within a specified directory. It can upload, delete, preview, and edit files through a web interface.

This fork continues maintenance after the upstream project announced its wind-down. The current priorities are vulnerability remediation, reproducible releases, workflow hardening, and triage of the inherited backlog.

## Security

Start with:

- [Security policy](SECURITY.md)
- [Security maintenance and deployment controls](docs/security-maintenance.md)
- [Security remediation plan](docs/security-remediation-plan.md)
- [Master security tracker](https://github.com/cryptnetworks/filebrowser/issues/2)
- [Critical signup-scope fix](https://github.com/cryptnetworks/filebrowser/issues/3)

Until the linked fixes ship:

- Keep self-signup disabled.
- Keep command execution, runners, and hooks disabled.
- Run the service unprivileged with the smallest possible mounted filesystem.
- Put it behind a trusted reverse proxy that provides TLS and independent authentication.
- Treat existing JWT sessions as non-revocable until expiry.
- Keep tested offline backups.

Do not interpret the existence of scanning workflows as proof that the current code is safe. CodeQL, dependency, OSV, and DAST results are inputs to review; the eleven inherited advisories with no patched upstream version remain tracked work.

## Documentation

Installation, configuration, and build documentation lives in [`docs`](docs). Fork-specific security operations are documented in [`docs/security-maintenance.md`](docs/security-maintenance.md).

[`CONTRIBUTING.md`](CONTRIBUTING.md) describes how to build and develop the project.

## License

Apache License 2.0 © File Browser Contributors
