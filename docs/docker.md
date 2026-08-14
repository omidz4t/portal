# Docker

Portal publishes a **public** image on GitHub Container Registry.

| | |
|--|--|
| Image | `ghcr.io/omidz4t/portal` |
| Package | [github.com/users/omidz4t/packages/container/package/portal](https://github.com/users/omidz4t/packages/container/package/portal) |
| Platforms | `linux/amd64`, `linux/arm64` |
| User | `portal` (uid 10001) |
| Default command | `serve` with `--config /etc/portal/config.yml --folder /var/lib/portal` |

The image contains `portal` and `deltachat-rpc-server` (v2.56). **Secrets stay outside the image.**

## Tags

CI on `main` pushes:

- `latest`
- full git SHA (`ghcr.io/omidz4t/portal:<sha>`)
- the value in `VERSION`

Pull requests only **build** the image; they do not push.

## Run

```bash
cp config.example.yml config.yml
cp .env.example .env
# edit TELEGRAM_BOT_TOKEN (and other secrets) in .env

docker pull ghcr.io/omidz4t/portal:latest

docker run --rm \
  -v "$PWD/config.yml:/etc/portal/config.yml:ro" \
  --env-file .env \
  -v portal-data:/var/lib/portal \
  ghcr.io/omidz4t/portal:latest \
  --config /etc/portal/config.yml --folder /var/lib/portal \
  init dcaccount:nine.testrun.org

docker run --rm \
  -v "$PWD/config.yml:/etc/portal/config.yml:ro" \
  --env-file .env \
  -v portal-data:/var/lib/portal \
  ghcr.io/omidz4t/portal:latest
```

Set `folder: /var/lib/portal` in `config.yml` so it matches the volume.

## Build locally

```bash
make docker
# → ghcr.io/omidz4t/portal:<VERSION> and :local
```

See also [installation.md](installation.md) and [security.md](security.md).
