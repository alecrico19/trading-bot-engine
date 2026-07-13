# Clean Restart — Fix Ctrl+C Issues

## Problem
Ctrl+C leaves zombie processes, port conflicts, and Go cache builds. Restarts aren't truly clean.

## Fix: A `restart.sh` script that nukes everything first

```bash
#!/bin/bash
# ~/trading-bot/scripts/restart.sh

echo "=== Killing all bot processes ==="
pkill -9 -f "go run.*main.go" 2>/dev/null
pkill -9 -f "launcher.sh" 2>/dev/null
pkill -9 -f "trading-bot" 2>/dev/null
pkill -9 -f "redis-server" 2>/dev/null
fuser -k 8080/tcp 2>/dev/null
sleep 2

echo "=== Cleaning ==="
export PATH=$HOME/go/bin:$HOME/go-tools/bin:$PATH
export GOROOT=$HOME/go
go clean -cache 2>/dev/null
rm -f $HOME/trading-bot/execution/trading.db*

echo "=== Starting ==="
$HOME/Desktop/start-trading-bot.sh
```

## Or: Add these lines to the TOP of `start-trading-bot.sh`

The existing launcher already has a cleanup step (`Cleaning stale ports`), but it's too gentle. Replace it with the aggressive kill above.

## Also: Add `go clean -cache` to the launcher

Right before `go run cmd/main.go`, add:
```bash
go clean -cache 2>/dev/null
```

This ensures every restart compiles from fresh source, not cached binaries.

## Files
- `scripts/restart.sh` — new file (or add to existing launcher)
- `scripts/launcher.sh` — add `go clean -cache` before engine start