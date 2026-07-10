# Systemd Service Fix

## Current State
- Service file exists at `~/.config/systemd/user/trading-bot.service`
- Status: **disabled** (won't auto-start on boot)
- Service type: `simple` — but launcher.sh backgrounds processes, which breaks systemd lifecycle tracking

## Problem
`launcher.sh` uses `&` to background engine and research. When systemd sends SIGTERM to the launcher, only the launcher dies — the child processes survive as orphans.

## Fix Options

### Option A: Restructure launcher as foreground (recommended)
- Launcher runs engine in foreground (no `&`)
- Research service as a separate systemd unit or launched via `ExecStartPost`
- Clean SIGTERM → clean shutdown of all children

### Option B: Use systemd to manage all services independently
- Separate `.service` files for Redis, engine, research, telegram
- `trading-bot.target` groups them all
- `Requires=` / `After=` for dependency ordering
- Most robust but more files

### Option C: Minimal — just enable + linger
- Enable the current service as-is
- Add `loginctl enable-linger alecr`
- Accept orphaned processes on restart (minor issue)

## Recommendation
Option A — fix the launcher to run engine in foreground. Research starts as `ExecStartPost`. Simplest, works with systemd, clean shutdown.

## Commands to Run
```bash
loginctl enable-linger alecr
systemctl --user daemon-reload
systemctl --user enable trading-bot
systemctl --user start trading-bot
```
