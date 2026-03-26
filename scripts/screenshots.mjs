/**
 * screenshots.mjs — capture screenshots of every Traefik GUI page with
 * realistic mock data injected via Playwright route interception.
 *
 * Prerequisites:
 *   cd scripts && npm install && npx playwright install chromium
 *
 * Usage:
 *   APP_URL=http://localhost:8888 node screenshots.mjs
 *
 * The app must be running (go run ../main.go, or docker run …).
 * All /api/* and /auth/* calls are intercepted — no real Traefik needed.
 *
 * Output: ../docs/screenshots/*.png
 */

import { chromium } from 'playwright'
import { mkdir } from 'fs/promises'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const OUT_DIR = path.join(__dirname, '..', 'docs', 'screenshots')
const APP_URL = process.env.APP_URL ?? 'http://localhost:8888'

// ── Mock data ────────────────────────────────────────────────────────────────

const now = new Date()
const daysFromNow = d => new Date(now.getTime() + d * 86400000).toISOString()

const MOCKS = {
  // Auth
  '/auth/check': { user: 'admin' },
  '/auth/login': { user: 'admin' },

  // Status
  '/api/status': { gui: 'ok', traefik: true },

  // Static config
  '/api/config': {
    traefikApiUrl: 'http://traefik:8080',
    paths: {
      staticConfig: '/etc/traefik/traefik.yml',
      dynamicDir: '/etc/traefik/dynamic',
      acmePath: '/etc/traefik/acme.json',
      staticConfigFound: true,
      dynamicDirFound: true,
      acmePathFound: true,
    },
    staticConfig: {
      api: { dashboard: true, insecure: false },
      entryPoints: {
        web: { address: ':80', http: { redirections: { entryPoint: { to: 'websecure', scheme: 'https', permanent: true } } } },
        websecure: { address: ':443', http: { tls: { certResolver: 'cf-dns' } } },
      },
      providers: {
        docker: { endpoint: 'unix:///var/run/docker.sock', exposedByDefault: false, network: 'traefik', watch: true },
        file: { directory: '/etc/traefik/dynamic', watch: true },
      },
      certificatesResolvers: {
        'cf-dns': { acme: { email: 'admin@example.com', storage: '/etc/traefik/acme.json', dnsChallenge: { provider: 'cloudflare', resolvers: ['1.1.1.1:53'] } } },
      },
      log: { level: 'INFO' },
      accessLog: { format: 'json' },
      global: { checkNewVersion: false, sendAnonymousUsage: false },
    },
    rawConfig: `api:\n  dashboard: true\n  insecure: false\n\nentryPoints:\n  web:\n    address: ":80"\n  websecure:\n    address: ":443"\n\nproviders:\n  docker:\n    exposedByDefault: false\n  file:\n    directory: /etc/traefik/dynamic\n    watch: true\n`,
  },

  // Traefik API proxy — overview
  '/api/traefik/api/overview': {
    http: {
      routers:     { total: 12, warnings: 0, errors: 1 },
      services:    { total: 11, warnings: 0, errors: 0 },
      middlewares: { total: 4,  warnings: 0, errors: 0 },
    },
  },

  // Traefik API proxy — routers
  '/api/traefik/api/http/routers': [
    { name: 'dashboard@file',    rule: 'Host(`traefik.example.com`)',    status: 'enabled', service: 'api@internal',   entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'file' },
    { name: 'whoami@docker',     rule: 'Host(`whoami.example.com`)',     status: 'enabled', service: 'whoami',         entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'docker' },
    { name: 'immich@file',       rule: 'Host(`photos.example.com`)',     status: 'enabled', service: 'immich',         entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'file' },
    { name: 'jellyfin@docker',   rule: 'Host(`media.example.com`)',      status: 'enabled', service: 'jellyfin',       entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'docker' },
    { name: 'nextcloud@file',    rule: 'Host(`cloud.example.com`)',      status: 'enabled', service: 'nextcloud',      entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'file' },
    { name: 'grafana@docker',    rule: 'Host(`grafana.example.com`)',    status: 'enabled', service: 'grafana',        entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'docker' },
    { name: 'vaultwarden@file',  rule: 'Host(`vault.example.com`)',      status: 'enabled', service: 'vaultwarden',    entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'file' },
    { name: 'portainer@docker',  rule: 'Host(`portainer.example.com`)', status: 'enabled', service: 'portainer',      entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'docker' },
    { name: 'uptime@docker',     rule: 'Host(`status.example.com`)',     status: 'enabled', service: 'uptime-kuma',    entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'docker' },
    { name: 'homeassistant@file',rule: 'Host(`home.example.com`)',       status: 'enabled', service: 'homeassistant',  entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'file' },
    { name: 'pve@file',          rule: 'Host(`pve.example.com`)',        status: 'enabled', service: 'proxmox',        entryPoints: ['websecure'], tls: { certResolver: 'cf-dns' }, provider: 'file' },
    { name: 'broken@docker',     rule: 'Host(`broken.example.com`)',     status: 'error',   service: 'broken',         entryPoints: ['websecure'], provider: 'docker' },
  ],

  // Certificates
  '/api/certificates': {
    available: true,
    certs: [
      { resolver: 'cf-dns', domain: 'traefik.example.com',     sans: [],                                               expiry: daysFromNow(87),  daysLeft: 87 },
      { resolver: 'cf-dns', domain: 'whoami.example.com',      sans: [],                                               expiry: daysFromNow(62),  daysLeft: 62 },
      { resolver: 'cf-dns', domain: 'photos.example.com',      sans: ['www.photos.example.com'],                        expiry: daysFromNow(21),  daysLeft: 21 },
      { resolver: 'cf-dns', domain: 'media.example.com',       sans: [],                                               expiry: daysFromNow(14),  daysLeft: 14 },
      { resolver: 'cf-dns', domain: 'cloud.example.com',       sans: ['dav.cloud.example.com', 'office.example.com'],  expiry: daysFromNow(90),  daysLeft: 90 },
      { resolver: 'cf-dns', domain: 'grafana.example.com',     sans: [],                                               expiry: daysFromNow(55),  daysLeft: 55 },
      { resolver: 'cf-dns', domain: 'vault.example.com',       sans: [],                                               expiry: daysFromNow(-3),  daysLeft: -3 },
      { resolver: 'cf-dns', domain: 'portainer.example.com',   sans: [],                                               expiry: daysFromNow(7),   daysLeft: 7  },
      { resolver: 'cf-dns', domain: 'status.example.com',      sans: [],                                               expiry: daysFromNow(73),  daysLeft: 73 },
      { resolver: 'cf-dns', domain: 'home.example.com',        sans: [],                                               expiry: daysFromNow(44),  daysLeft: 44 },
    ],
  },

  // Docker containers
  '/api/docker': {
    available: true,
    containers: [
      { id: 'a1b2c3d4e5f6', name: 'traefik',       image: 'traefik:v3.2',           state: 'running', enabled: false, traefikLabels: { 'traefik.enable': 'false' } },
      { id: 'b2c3d4e5f6a1', name: 'traefik-gui',   image: 'ghcr.io/depuytnl/traefik-gui:latest', state: 'running', enabled: true,  traefikLabels: { 'traefik.enable': 'true', 'traefik.http.routers.traefik-gui.rule': 'Host(`traefik-gui.example.com`)', 'traefik.http.routers.traefik-gui.entrypoints': 'websecure' } },
      { id: 'c3d4e5f6a1b2', name: 'whoami',        image: 'traefik/whoami:latest',   state: 'running', enabled: true,  traefikLabels: { 'traefik.enable': 'true', 'traefik.http.routers.whoami.rule': 'Host(`whoami.example.com`)', 'traefik.http.routers.whoami.entrypoints': 'websecure', 'traefik.http.routers.whoami.tls.certresolver': 'cf-dns' } },
      { id: 'd4e5f6a1b2c3', name: 'jellyfin',      image: 'jellyfin/jellyfin:latest',state: 'running', enabled: true,  traefikLabels: { 'traefik.enable': 'true', 'traefik.http.routers.jellyfin.rule': 'Host(`media.example.com`)' } },
      { id: 'e5f6a1b2c3d4', name: 'grafana',       image: 'grafana/grafana:latest',  state: 'running', enabled: true,  traefikLabels: { 'traefik.enable': 'true', 'traefik.http.routers.grafana.rule': 'Host(`grafana.example.com`)' } },
      { id: 'f6a1b2c3d4e5', name: 'portainer',     image: 'portainer/portainer-ce',  state: 'running', enabled: true,  traefikLabels: { 'traefik.enable': 'true', 'traefik.http.routers.portainer.rule': 'Host(`portainer.example.com`)' } },
      { id: 'a2b3c4d5e6f7', name: 'uptime-kuma',   image: 'louislam/uptime-kuma:1',  state: 'running', enabled: true,  traefikLabels: { 'traefik.enable': 'true', 'traefik.http.routers.uptime.rule': 'Host(`status.example.com`)' } },
      { id: 'b3c4d5e6f7a2', name: 'influxdb',      image: 'influxdb:2.7',            state: 'running', enabled: false, traefikLabels: {} },
      { id: 'c4d5e6f7a2b3', name: 'redis',         image: 'redis:7-alpine',          state: 'running', enabled: false, traefikLabels: {} },
      { id: 'd5e6f7a2b3c4', name: 'postgres',      image: 'postgres:16-alpine',      state: 'running', enabled: false, traefikLabels: {} },
    ],
  },

  // Dynamic config files
  '/api/dynamic': [
    { name: 'dashboard.yml',     active: true,  hostnames: ['traefik.example.com'],  backends: ['api@internal'],          certResolver: 'cf-dns', insecureSkipVerify: false, routerCount: 1, serviceCount: 0, middlewareCount: 0 },
    { name: 'immich.yml',        active: true,  hostnames: ['photos.example.com'],   backends: ['http://immich:2283'],     certResolver: 'cf-dns', insecureSkipVerify: false, routerCount: 1, serviceCount: 1, middlewareCount: 0 },
    { name: 'nextcloud.yml',     active: true,  hostnames: ['cloud.example.com'],    backends: ['http://nextcloud:80'],    certResolver: 'cf-dns', insecureSkipVerify: false, routerCount: 1, serviceCount: 1, middlewareCount: 1 },
    { name: 'vaultwarden.yml',   active: true,  hostnames: ['vault.example.com'],    backends: ['http://vaultwarden:80'],  certResolver: 'cf-dns', insecureSkipVerify: false, routerCount: 1, serviceCount: 1, middlewareCount: 0 },
    { name: 'homeassistant.yml', active: true,  hostnames: ['home.example.com'],     backends: ['http://homeassistant:8123'], certResolver: 'cf-dns', insecureSkipVerify: false, routerCount: 1, serviceCount: 1, middlewareCount: 0 },
    { name: 'pve.yml',           active: true,  hostnames: ['pve.example.com'],      backends: ['https://192.168.1.10:8006'], certResolver: 'cf-dns', insecureSkipVerify: true,  routerCount: 1, serviceCount: 1, middlewareCount: 0 },
    { name: 'mtls.yml',          active: true,  hostnames: [],                       backends: [],                         certResolver: '',       insecureSkipVerify: false, routerCount: 0, serviceCount: 0, middlewareCount: 0 },
  ],

  // Activity log
  '/api/accesslog': {
    available: true,
    entries: [
      { time: new Date(now - 1000).toISOString(),   method: 'GET',  host: 'photos.example.com',    path: '/api/assets',               status: 200, durationMs: 14,  clientIp: '10.0.0.5',  routerName: 'immich@file',       serviceAddr: 'http://immich:2283',      entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 8420 },
      { time: new Date(now - 3000).toISOString(),   method: 'GET',  host: 'cloud.example.com',     path: '/remote.php/dav/files',     status: 207, durationMs: 32,  clientIp: '10.0.0.5',  routerName: 'nextcloud@file',    serviceAddr: 'http://nextcloud:80',     entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 1204 },
      { time: new Date(now - 5000).toISOString(),   method: 'POST', host: 'vault.example.com',     path: '/api/ciphers',              status: 200, durationMs: 8,   clientIp: '10.0.0.12', routerName: 'vaultwarden@file',  serviceAddr: 'http://vaultwarden:80',   entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 612 },
      { time: new Date(now - 8000).toISOString(),   method: 'GET',  host: 'grafana.example.com',   path: '/api/dashboards/home',      status: 200, durationMs: 22,  clientIp: '10.0.0.5',  routerName: 'grafana@docker',    serviceAddr: 'http://172.18.0.5:3000',  entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 3841 },
      { time: new Date(now - 12000).toISOString(),  method: 'GET',  host: 'media.example.com',     path: '/Videos/stream/uuid-1234',  status: 206, durationMs: 3,   clientIp: '10.0.0.8',  routerName: 'jellyfin@docker',   serviceAddr: 'http://172.18.0.7:8096',  entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 1048576 },
      { time: new Date(now - 18000).toISOString(),  method: 'GET',  host: 'status.example.com',    path: '/api/status-page/heartbeat',status: 200, durationMs: 4,   clientIp: '10.0.0.2',  routerName: 'uptime@docker',     serviceAddr: 'http://172.18.0.9:3001',  entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 286 },
      { time: new Date(now - 25000).toISOString(),  method: 'PUT',  host: 'cloud.example.com',     path: '/remote.php/dav/files/doc', status: 201, durationMs: 44,  clientIp: '10.0.0.5',  routerName: 'nextcloud@file',    serviceAddr: 'http://nextcloud:80',     entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 0 },
      { time: new Date(now - 33000).toISOString(),  method: 'GET',  host: 'home.example.com',      path: '/api/states',               status: 200, durationMs: 11,  clientIp: '10.0.0.3',  routerName: 'homeassistant@file',serviceAddr: 'http://homeassistant:8123',entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 9120 },
      { time: new Date(now - 41000).toISOString(),  method: 'GET',  host: 'photos.example.com',    path: '/api/assets/thumbnail',     status: 200, durationMs: 7,   clientIp: '10.0.0.5',  routerName: 'immich@file',       serviceAddr: 'http://immich:2283',      entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 24320 },
      { time: new Date(now - 55000).toISOString(),  method: 'GET',  host: 'portainer.example.com', path: '/api/endpoints',            status: 200, durationMs: 6,   clientIp: '10.0.0.5',  routerName: 'portainer@docker',  serviceAddr: 'http://172.18.0.11:9000', entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 1536 },
      { time: new Date(now - 72000).toISOString(),  method: 'GET',  host: 'grafana.example.com',   path: '/api/datasources/proxy',    status: 502, durationMs: 5001,clientIp: '10.0.0.5',  routerName: 'grafana@docker',    serviceAddr: 'http://172.18.0.5:3000',  entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 0 },
      { time: new Date(now - 90000).toISOString(),  method: 'DELETE',host: 'cloud.example.com',    path: '/remote.php/dav/trash/file',status: 204, durationMs: 19,  clientIp: '10.0.0.5',  routerName: 'nextcloud@file',    serviceAddr: 'http://nextcloud:80',     entryPoint: 'websecure', tlsVersion: 'TLSv1.3', responseSize: 0 },
    ],
  },

  // mTLS
  '/api/mtls': {
    caExists: true,
    caExpires: daysFromNow(3650),
    applied: true,
    clients: [
      { id: 'cli_abc123', name: 'Work laptop',   issued: daysFromNow(-30), expires: daysFromNow(335) },
      { id: 'cli_def456', name: 'Home desktop',  issued: daysFromNow(-14), expires: daysFromNow(351) },
      { id: 'cli_ghi789', name: 'Mobile phone',  issued: daysFromNow(-7),  expires: daysFromNow(358) },
      { id: 'cli_jkl012', name: 'Tablet',        issued: daysFromNow(-2),  expires: daysFromNow(363) },
    ],
  },

  // Audit log
  '/api/audit': {
    entries: [
      { time: new Date(now - 60000).toISOString(),   user: 'admin', action: 'save static config',    detail: 'traefik.yml updated' },
      { time: new Date(now - 3600000).toISOString(), user: 'admin', action: 'save dynamic config',   detail: 'immich.yml updated' },
      { time: new Date(now - 7200000).toISOString(), user: 'admin', action: 'issue client cert',     detail: 'issued cert for "Tablet"' },
      { time: new Date(now - 86400000).toISOString(),user: 'admin', action: 'save dynamic config',   detail: 'nextcloud.yml updated' },
      { time: new Date(now - 90000000).toISOString(),user: 'admin', action: 'restarted container',   detail: 'traefik' },
      { time: new Date(now - 172800000).toISOString(),user: 'admin', action: 'apply mtls',           detail: 'mtls.yml written' },
      { time: new Date(now - 259200000).toISOString(),user: 'admin', action: 'generate CA',          detail: 'mTLS CA created' },
      { time: new Date(now - 345600000).toISOString(),user: 'admin', action: 'save static config',   detail: 'traefik.yml updated' },
    ],
  },
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function routeKey(url) {
  const u = new URL(url)
  // Strip query string for matching
  const pathname = u.pathname
  // Match /api/traefik/api/http/routers?per_page=... → key without query
  for (const key of Object.keys(MOCKS)) {
    if (pathname === key || pathname.startsWith(key + '?')) return key
  }
  return pathname
}

async function interceptAll(page) {
  await page.route('**/*', async (route) => {
    const url = route.request().url()
    const u = new URL(url)
    const key = routeKey(url)

    if (MOCKS[key] !== undefined) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(MOCKS[key]),
      })
      return
    }

    // Let non-API requests through (HTML, JS, CSS, fonts)
    if (!u.pathname.startsWith('/api/') && !u.pathname.startsWith('/auth/')) {
      await route.continue()
      return
    }

    // Unknown API — return empty 200 so the UI doesn't error
    await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  })
}

async function screenshot(page, name, url, { waitFor, extraDelay = 0 } = {}) {
  await page.goto(`${APP_URL}${url}`, { waitUntil: 'networkidle' })
  if (waitFor) await page.waitForSelector(waitFor, { timeout: 5000 }).catch(() => {})
  if (extraDelay) await page.waitForTimeout(extraDelay)
  const file = path.join(OUT_DIR, `${name}.png`)
  await page.screenshot({ path: file, fullPage: false })
  console.log(`  ✓ ${name}.png`)
}

// ── Main ─────────────────────────────────────────────────────────────────────

async function main() {
  await mkdir(OUT_DIR, { recursive: true })

  const browser = await chromium.launch()
  const page = await browser.newPage()
  await page.setViewportSize({ width: 1280, height: 800 })

  await interceptAll(page)

  console.log(`Capturing screenshots from ${APP_URL} → ${OUT_DIR}\n`)

  const pages = [
    { name: '01-dashboard',      url: '/',             waitFor: '.card' },
    { name: '02-static-config',  url: '/static',       waitFor: '.card' },
    { name: '03-dynamic-config', url: '/dynamic',      waitFor: '.card' },
    { name: '04-certificates',   url: '/certificates', waitFor: '.card' },
    { name: '05-docker-labels',  url: '/docker',       waitFor: '.card' },
    { name: '06-activity',       url: '/activity',     waitFor: 'table', extraDelay: 300 },
    { name: '07-mtls',           url: '/mtls',         waitFor: '.card' },
    { name: '08-audit-log',      url: '/audit',        waitFor: '.card' },
  ]

  for (const p of pages) {
    await screenshot(page, p.name, p.url, p)
  }

  await browser.close()
  console.log('\nDone. Screenshots saved to docs/screenshots/')
}

main().catch(err => { console.error(err); process.exit(1) })
