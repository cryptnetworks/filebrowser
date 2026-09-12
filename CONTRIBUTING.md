# Building File Browser

This fork accepts pull requests against `master`. Security hardening and regression fixes take priority. The upstream archive does not apply to this fork. See the [security policy](SECURITY.md) and [operations runbooks](docs/operations.md).

Owner: `cryptnetworks`. Review these instructions monthly and whenever CI or runtime requirements change.

## Project Structure

The backend side of the application is written in [Go](https://golang.org/), while the frontend (located on a subdirectory of the same name) is written in [Vue.js](https://vuejs.org/). Due to the tight coupling required by some features, basic knowledge of both Go and Vue.js is recommended.

* Learn Go: [https://github.com/golang/go/wiki/Learn](https://github.com/golang/go/wiki/Learn)
* Learn Vue.js: [https://vuejs.org/guide/introduction.html](https://vuejs.org/guide/introduction.html)

We encourage you to use git to manage your fork. To clone the main repository, just run:

```bash
git clone https://github.com/cryptnetworks/filebrowser
```

We use [Taskfile](https://taskfile.dev/) to manage the different processes (building, releasing, etc) automatically.

## Build

You can fully build the project in order to produce a binary by running:

```bash
task build
```

## Development

For development, there are a few things to have in mind.

### Frontend

We use [Node.js](https://nodejs.org/en/) on the frontend to manage the build process. Prepare the frontend environment:

```bash
# From the root of the repo, go to frontend/
cd frontend

# Install the dependencies
pnpm install
```

If you just want to develop the backend, you can create a static build of the frontend:

```bash
pnpm run build
```

If you want to develop the frontend, start a development server which watches for changes:

```bash
pnpm run dev
```

Please note that you need to access File Browser's interface through the development server of the frontend.

### Backend

First prepare the backend environment by downloading all required dependencies:

```bash
go mod download
```

You can now build or run File Browser as any other Go project:

```bash
# Build
go build

# Run
go run .
```

## Documentation

Documentation lives in [`docs`](docs) as plain Markdown and is no longer built into a site. The command line reference in [`docs/cli`](docs/cli) is generated from the commands themselves:

```bash
task docs:cli:generate
```

## Release

Release tagging is a publishing action. Complete the [release checklist](docs/operations.md#release) before running:

```bash
task release
```

## Translations

The Transifex integration stopped on 2026-09-01 and translations submitted there no longer reach this repository. Locale files live in [`frontend/src/i18n`](frontend/src/i18n) and can be edited directly.

## Authentication Provider

To build a new authentication provider, you need to implement the [Auther interface](https://github.com/filebrowser/filebrowser/blob/master/auth/auth.go), whose method will be called on the login page after the user has submitted their login data.

```go
// Auther is the authentication interface.
type Auther interface {
    // Auth is called to authenticate a request.
    Auth(r *http.Request, s users.Store, settings *settings.Settings, server *settings.Server) (*users.User, error)
    LoginPage() bool
}
```

After implementing the interface you should:

1. Add it to [`auth` directory](https://github.com/filebrowser/filebrowser/blob/master/auth).
2. Add it to the [configuration parser](https://github.com/filebrowser/filebrowser/blob/master/cmd/config.go) for the CLI.
3. Add it to the [`authBackend.Get`](https://github.com/filebrowser/filebrowser/blob/master/storage/bolt/auth.go).

If you need to add more flags, please update the function `addConfigFlags`.


## Runtime and verification

Use Go 1.26.6 (see `go.mod`), Node.js 24, and pnpm 10.33.4 (see
`frontend/package.json`). Install pnpm with `npm install --global pnpm@10.33.4`
if it is unavailable. The backend embeds the Vue frontend from `frontend/dist`.
Linux runners provide CI; release build targets are defined in `.goreleaser.yml`.
GHCR supports Linux amd64/arm64, plus arm/v7 for the standard image.

Run from the repository root:

```sh
go mod download
pnpm --dir frontend install --frozen-lockfile
gofmt -l auth cmd files http users settings storage runner img fileutils diskcache rules
pnpm --dir frontend run lint
pnpm --dir frontend run test
pnpm --dir frontend run build
go test --race ./...
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
"$(go env GOPATH)/bin/golangci-lint" run
go build -o filebrowser .
git diff --check
```

The `gofmt -l` check should print no paths; use `gofmt -w` on any changed Go
files that need formatting.
Use `pnpm --dir frontend exec prettier --write <file>` for changed frontend
files. `pnpm --dir frontend run typecheck` runs type checking alone; `build`
already includes it. HTTP integration tests run within `go test --race ./...`;
there is no separate integration-test command. `task build` is the CI production
build entry point and combines dependency installation, frontend build, and
backend build with version metadata.

## Contribution and compatibility rules

Use a focused branch from `master` and a pull request; do not push directly to
`master`. Link the relevant issue and describe behavior changes, migration,
and validation. Use the existing `security` label for disclosed security work;
do not invent milestone dates. Privately reported vulnerabilities follow
`SECURITY.md` before public discussion.

Preserve API and persisted-configuration compatibility unless a documented
security change requires otherwise. Persisted data uses Bolt; there is no
standalone database-migration command. Schema changes need backward-read tests,
an explicit upgrade strategy, and backup/restore instructions. Never run two
processes against the same database.

Definition of done: the local checks above pass, applicable security scans pass,
and regressions have tests. Protected checks currently include `Lint Frontend`,
`Test Frontend`, `Lint Backend`, `Test`, `Build`, and CodeQL `Analyze (actions)`,
`Analyze (go)`, `Analyze (javascript-typescript)`. Keep these names stable.
Dependency review and OSV also require review even though they are not currently
listed in branch protection. Publication has additional requirements in the
release runbook. A passing build alone does not establish a security-ready release.
