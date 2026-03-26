# Traefik GUI

A lightweight web-based management interface for [Traefik](https://traefik.io/) reverse proxy. Manage your static configuration, dynamic file configs, TLS certificates, mTLS, Docker labels, and access logs — all from a clean dark UI, deployed as a single Docker container.

> Built with [Claude Code](https://claude.ai/code) by Anthropic.

---

## Features

- **Dashboard** — live Traefik connection status and resolved config paths
- **Static Config editor** — form-based editor for `traefik.yml` (entry points, providers, certificate resolvers, access logging, Docker integration)
- **Dynamic Config editor** — create, edit, and delete dynamic file provider configs (routers, services, middlewares, TLS options)
- **Certificates** — view ACME certificates from `acme.json`, expiry badges, per-domain detail
- **Docker Labels inspector** — browse running containers and their Traefik labels
- **Activity log** — live access log tail from the Traefik container, with pause/clear and a toggle to hide GUI traffic
- **mTLS** — generate a CA, issue and revoke client certificates (PKCS#12 + PEM ZIP download), apply the TLS option to Traefik automatically
- **Audit Log** — record of all changes made through the GUI
- **Authentication** — simple username/password login protecting the entire interface

---

## Quick Start

```yaml
# docker-compose.yml
services:
  traefik-gui:
    image: ghcr.io/deputynl/traefik-gui:latest
    container_name: traefik-gui
    restart: unless-stopped
    environment:
      TRAEFIK_CONFIG_PATH: /etc/traefik/traefik.yml
      TRAEFIK_API_URL: http://traefik:8080
      TRAEFIK_GUI_USER: ${TRAEFIK_GUI_USER:-admin}
      TRAEFIK_GUI_PASSWORD: ${TRAEFIK_GUI_PASSWORD:-admin}
      TRAEFIK_ACME_PATH: /etc/traefik/acme.json
      TRAEFIK_CONTAINER_NAME: traefik
    volumes:
      - /etc/traefik:/etc/traefik          # read+write for config changes
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - traefik
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.traefik-gui.rule=Host(`traefik-gui.yourdomain.org`)"
      - "traefik.http.routers.traefik-gui.entrypoints=websecure"
      - "traefik.http.routers.traefik-gui.tls.certresolver=cf-dns"
      - "traefik.http.services.traefik-gui.loadbalancer.server.port=8888"

networks:
  traefik:
    external: true
```

The GUI is available on port `8888`. Default credentials are `admin` / `admin` — **change these via environment variables before exposing publicly.**

---

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `TRAEFIK_GUI_PORT` | `8888` | Port the GUI listens on |
| `TRAEFIK_CONFIG_PATH` | `/etc/traefik/traefik.yml` | Path to Traefik static config |
| `TRAEFIK_API_URL` | `http://localhost:8080` | Traefik API base URL |
| `TRAEFIK_GUI_USER` | `admin` | GUI login username |
| `TRAEFIK_GUI_PASSWORD` | `admin` | GUI login password |
| `TRAEFIK_ACME_PATH` | *(derived from config path)* | Override path to `acme.json` (useful when paths differ between containers) |
| `TRAEFIK_CONTAINER_NAME` | `traefik` | Docker container name for access log streaming |

---

## mTLS Setup

The mTLS page guides you through three steps:

1. **Generate CA** — creates a self-signed Certificate Authority stored in the container's data directory
2. **Apply to Traefik** — writes `mtls.yml` to your dynamic config directory, defining the `mtls` TLS option with `clientAuth.caFiles` pointing to the CA cert
3. **Issue client certificates** — generates a signed client cert/key pair, bundles them as a PKCS#12 (`.p12`) and PEM ZIP for import into browsers or HTTP clients

To use mTLS on an entry point, set `tls.options: mtls@file` on the router in your dynamic config.

---

## Architecture

- **Backend**: Go 1.22, `net/http` stdlib, no heavy framework
- **Frontend**: Vue 3 + Vite + TypeScript + Tailwind CSS
- **Deployment**: `//go:embed` bundles the Vite build into the Go binary — single self-contained binary, single container
- **Image**: multi-arch (`linux/amd64`, `linux/arm64`), based on Alpine, ~25 MB

```
traefik-gui/
├── main.go
├── internal/
│   ├── accesslog/    # Docker log streaming + parsing
│   ├── api/          # HTTP server, route registration, all handlers
│   ├── audit/        # Append-only audit log
│   ├── auth/         # Session-based authentication
│   ├── config/       # Env var loading + path resolution
│   ├── docker/       # Docker socket client (containers, restart)
│   ├── mtls/         # CA + client cert generation, PKCS#12 bundling
│   └── traefik/      # traefik.yml types + parser
└── web/src/
    ├── views/        # One Vue component per page
    ├── stores/       # Pinia stores (auth, config, certs)
    └── components/   # Shared UI components
```

---

## Development

```bash
# Frontend dev server (proxies /api/* to localhost:8888)
cd web && npm install && npm run dev

# Backend
go run main.go

# Production build + push to registry (multi-arch)
make release
```

---

## License

MIT
