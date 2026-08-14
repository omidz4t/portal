# Portal — Telegram ↔ Delta Chat bridge
# Runtime needs a mounted config + env. Do not bake tokens into the image.
#
#   docker build -t ghcr.io/omidz4t/portal:local --build-arg VERSION=dev .
#   docker run --rm -v "$PWD/config.yml:/etc/portal/config.yml:ro" \
#     --env-file .env ghcr.io/omidz4t/portal:local serve

# syntax=docker/dockerfile:1

FROM golang:bookworm AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/portal ./cmd/portal

FROM debian:bookworm-slim
ARG TARGETARCH
ARG RPC_VERSION=v2.56.0
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl \
	&& rm -rf /var/lib/apt/lists/* \
	&& case "${TARGETARCH}" in \
		amd64) rpc_arch=x86_64 ;; \
		arm64) rpc_arch=aarch64 ;; \
		*) echo "unsupported TARGETARCH=${TARGETARCH}" >&2; exit 1 ;; \
	esac \
	&& curl -fsSL -o /usr/local/bin/deltachat-rpc-server \
		"https://github.com/chatmail/core/releases/download/${RPC_VERSION}/deltachat-rpc-server-${rpc_arch}-linux" \
	&& chmod 0755 /usr/local/bin/deltachat-rpc-server \
	&& useradd --system --home /var/lib/portal --shell /usr/sbin/nologin --uid 10001 portal \
	&& mkdir -p /var/lib/portal /etc/portal \
	&& chown -R portal:portal /var/lib/portal /etc/portal

COPY --from=build /out/portal /usr/bin/portal
COPY config.example.yml /usr/share/portal/config.example.yml

USER portal
WORKDIR /var/lib/portal
VOLUME ["/var/lib/portal"]
ENTRYPOINT ["/usr/bin/portal"]
CMD ["--config", "/etc/portal/config.yml", "--folder", "/var/lib/portal", "serve"]
