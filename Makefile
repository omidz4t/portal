.PHONY: tidy build build-release build-release-all test test-proxy vet run init serve help clean config \
	version version-dry release-tag patch minor major run-landing

# Project: TGPORTAL
BINARY := tgportal
PKG    := ./cmd/tgportal
CMD    := ./$(BINARY)
CONFIG ?= config.yml
DIST   ?= dist
# Prefer VERSION file (semver), then override with make VERSION=…
VERSION_FILE := $(shell test -f VERSION && tr -d '[:space:]' < VERSION || echo dev)
VERSION ?= $(VERSION_FILE)
GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Optional: pass QR/account as make init QR=dcaccount:...
QR     ?=

# Static, stripped binary with version + embedded assets (see internal/assets).
LDFLAGS := -s -w -X main.version=$(VERSION)

tidy:
	go mod tidy

# Create local config/env from examples if missing
config:
	@test -f config.yml || cp config.example.yml config.yml
	@test -f .env || cp .env.example .env
	@echo "local config ready: config.yml, .env"

vet:
	go vet ./...

# Dev build at repo root (assets embedded; still CGO-free).
build: tidy
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) $(PKG)

# All-in-one release binary → dist/
# Single static executable: no CGO, stripped, version stamped, brand assets embedded.
# Runtime still needs: config.yml + .env (secrets) and deltachat-rpc-server on PATH.
build-release: tidy
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(DIST)/$(BINARY) $(PKG)
	@cp -f config.example.yml $(DIST)/config.example.yml
	@test -f .env.example && cp -f .env.example $(DIST)/.env.example || true
	@printf '%s\n' "$(VERSION)" > $(DIST)/VERSION
	@( cd $(DIST) && sha256sum $(BINARY) > $(BINARY).sha256 )
	@echo "release ready:"
	@ls -lh $(DIST)/$(BINARY) $(DIST)/$(BINARY).sha256
	@$(DIST)/$(BINARY) --version

# Cross-compile all-in-one archives for common platforms into dist/
build-release-all: tidy
	@mkdir -p $(DIST)
	@set -e; \
	for t in "linux amd64" "linux arm64" "darwin amd64" "darwin arm64" "windows amd64"; do \
	  set -- $$t; goos=$$1; goarch=$$2; \
	  ext=""; if [ "$$goos" = "windows" ]; then ext=".exe"; fi; \
	  name="$(BINARY)_$(VERSION)_$${goos}_$${goarch}"; \
	  out="$(DIST)/$${name}$${ext}"; \
	  echo "→ $$out"; \
	  CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch go build -trimpath -ldflags="$(LDFLAGS)" -o "$$out" $(PKG); \
	  ( cd $(DIST) && \
	    if [ -n "$$ext" ]; then zip -9 "$${name}.zip" "$${name}$${ext}" && rm -f "$${name}$${ext}"; \
	    else tar -czf "$${name}.tar.gz" "$${name}" && rm -f "$${name}"; fi ); \
	done; \
	( cd $(DIST) && sha256sum * > checksums.txt 2>/dev/null || sha256sum *.* > checksums.txt ); \
	ls -lh $(DIST)

test:
	go test ./...

# Proxy unit tests (in-process SOCKS5 + HTTP reverse proxy)
test-proxy:
	go test ./internal/proxy/ -count=1 -v

run: build
	$(CMD) --config $(CONFIG) $(ARGS)

# Example: make init QR=dcaccount:nine.testrun.org
init: build
	@test -n "$(QR)" || (echo "usage: make init QR=dcaccount:..." && exit 1)
	$(CMD) --config $(CONFIG) init $(QR)

serve: build
	$(CMD) --config $(CONFIG) serve

# SvelteKit landing site (./landing). Installs npm deps if needed.
LANDING_PORT ?= 5173
run-landing:
	@test -d landing || (echo "missing landing/" && exit 1)
	@test -d landing/node_modules || (cd landing && npm install)
	cd landing && npm run dev -- --host 127.0.0.1 --port $(LANDING_PORT)

help: build
	$(CMD) --help

# Conventional-commit version bump (no extra deps):
#   fix: → patch (0.0.X)   feat: → minor (0.X.0)
version:
	@bash scripts/bump-version.sh

version-dry:
	@bash scripts/bump-version.sh --dry-run

# Explicit semver bumps (ignore commit history)
patch:
	@bash scripts/bump-version.sh --bump patch

minor:
	@bash scripts/bump-version.sh --bump minor

major:
	@bash scripts/bump-version.sh --bump major

# Bump from commits, commit VERSION, create tag vX.Y.Z (push yourself)
release-tag:
	@bash scripts/bump-version.sh --commit --tag

clean:
	rm -f $(BINARY)
	rm -rf dist/
	rm -rf landing/.svelte-kit landing/build
