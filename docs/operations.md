# Operations and release runbooks

Owner: `cryptnetworks`. Review monthly, after each security merge, and before
release. Contributor commands are in [CONTRIBUTING](../CONTRIBUTING.md).
The [security policy](../SECURITY.md) defines support and response targets.

## Backup, upgrade, and rollback

1. Record the deployed commit/image digest, startup flags, proxy configuration,
   filesystem ownership, and database path. Restrict backups to operators:
   the Bolt database and configuration exports contain authentication material.
2. Stop File Browser and stop external writers to the served tree. Back up the
   database, configuration, and served files together to protected offline storage.
   A live copy of a changing database is not an application-consistent backup.
3. Restore the backup into an isolated test environment using the old image.
   Verify login, scope isolation, a file download hash, and recovery of an upload.
4. Read the candidate's migration notes; test it on a separate restored copy.
   For this proxy change, old configurations default `auth.autoProvision` to
   false. Existing accounts work; unknown proxy users are denied until an
   operator explicitly enables provisioning. No database rewrite is required.
5. Deploy the candidate by immutable digest, preserving non-root ownership and
   minimal mounts. Keep signup, execution, and external symlink following disabled
   unless their separately documented risks have been accepted.
6. Verify health, login, uploads/download hashes, and cross-user access denial.
   If checks fail, stop the service and restore the matching old database,
   configuration, files, and image. Preserve post-upgrade writes separately for
   reconciliation; restoring a backup can lose changes made after that backup.

## Incident response

1. Restrict ingress at the proxy and stop uploads/execution. If necessary stop
   File Browser. Preserve logs and a protected copy of the database and files.
2. Record the affected commit, time window, scopes, exposed shares, and suspected
   entry point. Do not publish credentials, tokens, or private file contents.
3. Contact the maintainer using [private reporting](../SECURITY.md#reporting-a-vulnerability).
   Revoke exposed credentials at the identity provider and rotate affected secrets.
   File Browser JWT logout is not reliable revocation: keep ingress restricted
   until stolen tokens expire or a separately tested signing-key rotation forces
   all sessions out. Deleting a browser cookie alone is insufficient.
4. Rebuild from a reviewed commit and restore known-good data after closing the
   entry point. Recheck scope boundaries and share access before reopening.
5. Record root cause, affected versions, containment, recovery checks, and follow-up
   issues. Coordinate disclosure and credit through the security policy.

## Scanner triage

Download OSV JSON/SARIF from the exact commit's workflow run. Export open
Dependabot and code-scanning alerts separately; zero GitHub alerts does not mean
OSV has no residual findings. Normalize CVE/GHSA/GO aliases, identify the direct
or transitive package, and record runtime versus build/dev reachability.

Patch compatible reachable high/critical findings first. For each exception,
record advisory IDs, package/version, reason, regression evidence, owner,
compensating control, tracker, and expiry. Re-review on code/dependency changes
or expiry, whichever comes first. Do not add scanner ignores merely to obtain
a green check. Scanner failures or malformed/missing reports are failures to
obtain evidence, not clean results. See the [current baseline](security-remediation-plan.md).

## Release

1. Select a reviewed commit on protected `master`. Run all contributor checks and
   relevant DAST against an ephemeral candidate instance. Confirm CodeQL,
   dependency review, OSV, and secret-scanning evidence for that commit.
2. Resolve all critical/high findings or record reviewed, time-limited acceptance.
   Confirm the master security tracker, migration notes, and supported-version
   table accurately describe what the release fixes and what remains.
3. Before tagging, audit destinations and credentials. The legacy GoReleaser
   configuration still names upstream Docker Hub repositories and uses `GH_PAT`;
   it must be migrated before using that path for a fork release. GHCR publishing
   is separate and currently runs on master pushes and tags without waiting for CI.
   Treat rolling images as development builds.
4. Require binary/container SBOMs, checksums, signatures, provenance, and scan
   summaries before advertising a security-ready release. GHCR BuildKit creates
   SBOM/provenance attestations; that alone is not a signed release. Complete
   [#13](https://github.com/cryptnetworks/filebrowser/issues/13) for end-to-end
   signing and verification. Do not claim a signature exists until verified.
5. Only after these gates, use the reviewed release procedure to generate CLI
   docs/changelog and tag. `task release` commits and tags; pushing a version tag
   starts publication. It is not a dry-run verification command.
6. Download the published artifacts, compare their checksums, verify signatures
   against the documented fork identity, and inspect provenance for the exact
   source commit and workflow. Test a clean install and the backup/upgrade/rollback
   sequence before updating supported versions in `SECURITY.md`.

## CI maintenance

Keep required-check names stable. Review pinned action updates and their release
notes; verify the full commit SHA belongs to the official repository. Record
before/after workflow duration and summed job time using the same event type and
runner class. Separate cancelled/failed runs from successful-run comparisons.
The first-batch changes reuse the tag build's frontend artifact in the release
job and add `Test Frontend` as a release dependency. PR installs remain separate
so lint, tests, and build keep independent required-check results.
