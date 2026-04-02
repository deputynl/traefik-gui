# Traefik GUI

A lightweight, single-container web GUI for managing Traefik reverse proxy — built for self-hosters who want to expose services both on their LAN and to the internet, with proper mTLS support and without touching YAML files.

> Built because getting mTLS working in Traefik is painful. This makes it easy. Thanks for your help [Claude Code](https://claude.ai/code) by Anthropic.

---

## Screenshots

| Dashboard | Certificates |
|-----------|--------------|
| ![Dashboard](https://raw.githubusercontent.com/depuytnl/traefik-gui/main/docs/img/01-dashboard.png) | ![Certificates](https://raw.githubusercontent.com/depuytnl/traefik-gui/main/docs/img/04-certificates.png) |

| Dynamic Config | Activity Log |
|----------------|--------------|
| ![Dynamic Config](https://raw.githubusercontent.com/depuytnl/traefik-gui/main/docs/img/03-dynamic-config.png) | ![Activity](https://raw.githubusercontent.com/depuytnl/traefik-gui/main/docs/img/06-activity.png) |

| mTLS | Audit Log |
|------|-----------|
| ![mTLS](https://raw.githubusercontent.com/depuytnl/traefik-gui/main/docs/img/07-mtls.png) | ![Audit Log](https://raw.githubusercontent.com/depuytnl/traefik-gui/main/docs/img/08-audit-log.png) |

---

## Features

- **mTLS management** — create and manage client certificates from the UI; delivered as a zip with all required formats and a README for easy installation on any device- **Dashboard** — live Traefik connection status and resolved config paths
- **Static Config editor** — form-based editor for `traefik.yml` (entry points, providers, certificate resolvers, access logging, Docker integration)
- **Dynamic Config editor** — create, edit, and delete dynamic file provider configs (routers, services, middlewares, TLS options)
- **Certificates** — view ACME certificates from `acme.json`, expiry badges, per-domain detail
- **Docker Labels inspector** — browse running containers and their Traefik labels
- **Activity log** — live access log tail from the Traefik container, with pause/clear and a toggle to hide GUI traffic
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
      - traefik-mtls:/etc/traefik/mtls     # CA, CA key + client certs, persisted across restarts
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

volumes:
  traefik-mtls:
```

The GUI is available on port `8888`. Default credentials are `admin` / `admin` — **change these via environment variables before exposing publicly.**

---

## Who is this for?

This tool makes opinionated assumptions: you're running Traefik with file-based dynamic config (not Kubernetes), you want to expose services on both LAN and internet, and you manage your own certificates. If that matches your homelab setup, this GUI will feel polished and complete. If you need a more flexible config editor, check out the alternatives below.

## Alternatives

- [Mantrae](https://github.com/MizuchiLabs/mantrae) — more flexible, good for general dynamic config management
- [Rahn-IT traefik-gui](https://github.com/Rahn-IT/traefik-gui) — simple routes editor
- [Traefikr](https://github.com/allfro/traefikr) — schema-validated forms for all Traefik config types

This project is more opinionated than those — if the assumptions fit your setup, the experience is more polished. If they don't, one of the above may suit you better.

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

1. **Generate CA** — creates a self-signed Certificate Authority; all files (`ca.crt`, `ca.key`) are stored in the `traefik-mtls` volume at `/etc/traefik/mtls/`
2. **Apply to Traefik** — writes `mtls.yml` to your dynamic config directory, defining the `default` TLS option with `clientAuth.caFiles` pointing to the CA cert; naming it `default` makes mTLS apply to all TLS connections automatically
3. **Issue client certificates** — generates a signed client cert/key pair, bundles them as a PKCS#12 (`.p12`) and PEM ZIP for import into browsers or HTTP clients; certificates are persisted in the `traefik-mtls` volume

Mount the `traefik-mtls` volume in your Traefik container so it can read `ca.crt`:

```yaml
# In your Traefik service
volumes:
  - traefik-mtls:/etc/traefik/mtls:ro
```

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
