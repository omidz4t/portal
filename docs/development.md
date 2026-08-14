# Development

## Prerequisites

See [installation.md](installation.md). Use `make` for routine tasks ([AGENTS.md](../AGENTS.md)).

## Build & test

```bash
make tidy
make build
make test
```

Binary: `./portal` (gitignored).

## CI / GitHub Actions

| Workflow | File | Trigger |
|----------|------|---------|
| CI | `.github/workflows/ci.yml` | push & PR to `main` |
| Deploy | `.github/workflows/deploy.yml` | after a version bump, or **Actions → Deploy production** |

CI tests and, on `main`, may tag a release. The deploy job then streams the linux/amd64 binary to the host. SSH secrets live in the **production** environment (`DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`, `DEPLOY_KNOWN_HOSTS`). They are **not** available to pull requests. Deploy runs only on `push` to `main` after a version bump, or via **Actions → Deploy production** (write access required). The deploy key is forced-command only (`portal-swap`).

### Version bump (no dependencies)

`scripts/bump-version.sh` — local or CI.

| Commits since last `v*` tag | Bump |
|----------------------------|------|
| `fix:` / `perf:` / `revert:` | **patch** → `0.0.X` |
| `feat:` | **minor** → `0.X.0` |
| `feat!:` / `BREAKING CHANGE` | **major** → `X.0.0` |
| `chore` / `docs` / `ci` / … | no bump |

```bash
make version-dry          # preview from commits
make version              # write VERSION only
make patch                # 0.0.X + changelog + commit + tag
make minor                # 0.X.0 + changelog + commit + tag
make major                # X.0.0 + changelog + commit + tag
make release-tag          # bump from conventional commits + changelog + commit + tag
git push origin HEAD --tags
```

`VERSION` is stamped into the binary (`-X main.version=…`).

### Publishing a release

Push to `main` with `feat:` / `fix:` commits, or tag locally:

```bash
make release-tag
git push origin HEAD --tags
```

The tag workflow tests, cross-builds, and uploads archives to the GitHub Release.  

Requires `contents: write` (default `GITHUB_TOKEN` on the repo).

## Run locally

```bash
make config
# set TELEGRAM_BOT_TOKEN in .env
make init QR=dcaccount:nine.testrun.org   # once
make serve
```

Override config path:

```bash
make serve CONFIG=./config.yml
```

## Project conventions

### Layout

- Application logic lives under `internal/…`  
- Only thin `main` under `cmd/portal`  
- Templates: `*.example*`; locals gitignored  

### Commits

This repo uses **signed** [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat: add unpair command
fix: avoid blocking Bot.Run on send
docs: document pairing flow
```

Always sign:

```bash
git commit -S -m "docs: …"
```

Commit after each logical change; never commit secrets, binaries, or `./data`.

### Code style

- Prefer `dc.Session` for all Delta Chat RPC  
- Keep Telegram handlers thin; put policy in `bridge` / `store`  
- Add tests for pure logic (`store`, `dc` invite parse, session mutex)  

## Useful CLI during development

```bash
./portal list
./portal link
./portal config addr          # if configured
./portal --folder ./data serve
```

## Debugging tips

| Issue | Approach |
|-------|----------|
| RPC / core version | Align `deltachat-rpc-server` with `go.mod` |
| Pairing | Inspect `data/tgportal.db` (`pairs` table) |
| TG not starting | Token, network, logs for `authorized as @…` |
| Media silent fail | Logs `bridged to Delta Chat` / `bridge error` |

SQLite peek (example):

```bash
sqlite3 data/tgportal.db 'SELECT id, code, telegram_user_id, status, dc_chat_id FROM pairs;'
```

## Adding a feature checklist

1. Config keys in `internal/config` + `config.example.yml`  
2. Behavior under the right `internal/` package  
3. Docs under `docs/` + README pointer if user-facing  
4. `make test` / `make build`  
5. Signed conventional commit  

## Module path

```text
github.com/omidz4t/portal
```

Adjust `go.mod` / import paths if you fork under another module path.
