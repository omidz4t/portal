# Development

## Prerequisites

See [installation.md](installation.md). Use `make` for routine tasks ([AGENTS.md](../AGENTS.md)).

## Build & test

```bash
make tidy
make build
make test
```

Binary: `./tgportal` (gitignored).

## CI / GitHub Actions

| Workflow | File | Trigger |
|----------|------|---------|
| CI | `.github/workflows/ci.yml` | push & PR to `main` |
| Release | `.github/workflows/release.yml` | tags `v*` |

CI runs `go mod tidy` check, `go vet`, `go test`, and cross-builds. After a green push to `main`, it bumps `VERSION` from conventional commits, prepends `CHANGELOG.md`, tags `vX.Y.Z`, and publishes the GitHub Release (token tag pushes do not start a second workflow). Commits starting with `chore(release):` are skipped. Manual tags still use `.github/workflows/release.yml`.

### Version bump (no dependencies)

`scripts/bump-version.sh` — local or CI.

| Commits since last `v*` tag | Bump |
|----------------------------|------|
| `fix:` / `perf:` / `revert:` | **patch** → `0.0.X` |
| `feat:` | **minor** → `0.X.0` |
| `feat!:` / `BREAKING CHANGE` | **major** → `X.0.0` |
| `chore` / `docs` / `ci` / … | no bump |

```bash
make version-dry          # preview
make version              # write VERSION only
make patch                # force patch
make release-tag          # bump + changelog + commit + tag
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
- Only thin `main` under `cmd/tgportal`  
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
./tgportal list
./tgportal link
./tgportal config addr          # if configured
./tgportal --folder ./data serve
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
github.com/themadorg/tgportal
```

Adjust `go.mod` / import paths if you fork under another module path.
