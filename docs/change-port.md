# Change Port from 8080 to 8420

## Why
SearXNG is using 8080 — that's what caused all the port conflicts, not stale Python processes. The engine never properly bound to 8080 because SearXNG was already there.

## What to Change

| File | Change |
|------|--------|
| `scripts/launcher.sh` | `--api 8080` → `--api 8420` |
| `scripts/launcher.sh` | Status ticker URL: `8080` → `8420` |
| `scripts/start.sh` | `8080` → `8420` |
| `cmd/main.go` | Default API port comment (keep flag dynamic) |
| `telegram/bot.py` | Default `API_URL` — this uses env var, keep as-is (user sets it) |
| `config/config.yaml` | No port — CLI flag only |

## Dashboard (no change needed)
Uses `window.location.origin` — auto-detects the port from the browser URL. No hardcoded port.

## How Many Places
5-6 string replacements. ~8 total changes across 3 files.

## New Default
**8420** — uncommon, not used by SearXNG, not a standard service port.

## Tailscale
Tailscale forwards whatever port the engine binds to. No change needed — just access `http://100.x.x.x:8420` on your phone.
